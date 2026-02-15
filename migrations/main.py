import sqlite3
import psycopg2
from psycopg2.extras import execute_values
from concurrent.futures import ThreadPoolExecutor, as_completed
from tqdm import tqdm
import argparse
import time

# =======================
# CONFIGURATION
# =======================
SQLITE_DB = "/data/Spotify/backup/spotify_clean.sqlite3"

PG = dict(
    dbname="bluppi_music",
    user="ethernode",
    password="password",
    host="localhost",
    port=5432,
)

# OPTIMIZED SETTINGS
# Smaller ranges prevents SQLite stalls during complex joins
# 10k range / 2.5k batch is the sweet spot for 16GB RAM + SSD
RANGE_SIZE = 100_000   
BATCH_SIZE = 10_000   
WORKERS = 10

# =======================
# DATABASE CONNECTIONS
# =======================
def sqlite_conn():
    """Optimized SQLite connection for read-heavy operations."""
    c = sqlite3.connect(SQLITE_DB)
    c.row_factory = sqlite3.Row
    c.execute("PRAGMA mmap_size=1073741824;") # 1GB mmap
    c.execute("PRAGMA temp_store=MEMORY;")
    c.execute("PRAGMA cache_size=-128000;")   # 128MB cache
    c.execute("PRAGMA journal_mode=OFF;")     # No journaling needed for reads
    c.execute("PRAGMA synchronous=OFF;")
    return c, c.cursor()

def pg_conn():
    """Optimized Postgres connection for bulk inserts."""
    c = psycopg2.connect(**PG)
    cur = c.cursor()
    cur.execute("SET synchronous_commit = OFF")
    cur.execute("SET work_mem = '512MB'")
    cur.execute("SET maintenance_work_mem = '1GB'")
    return c, cur

# =======================
# CHECKPOINT SYSTEM
# =======================
def init_checkpoints():
    """Create the checkpoint table to track progress."""
    pconn, pcur = pg_conn()
    pcur.execute("""
        CREATE TABLE IF NOT EXISTS etl_checkpoints (
            table_name TEXT NOT NULL,
            start_rowid BIGINT NOT NULL,
            end_rowid BIGINT NOT NULL,
            status TEXT DEFAULT 'pending',
            updated_at TIMESTAMPTZ DEFAULT now(),
            PRIMARY KEY (table_name, start_rowid, end_rowid)
        );
    """)
    pconn.commit()
    pconn.close()

def get_pending_ranges(table_name, row_min, row_max):
    """Generate all ranges and filter out the ones already completed."""
    pconn, pcur = pg_conn()
    
    # Get completed chunks
    pcur.execute("""
        SELECT start_rowid, end_rowid 
        FROM etl_checkpoints 
        WHERE table_name = %s AND status = 'done'
    """, (table_name,))
    
    completed = set((r[0], r[1]) for r in pcur.fetchall())
    pconn.close()
    
    # Generate all theoretical ranges
    all_ranges = []
    curr = row_min
    while curr <= row_max:
        end = min(curr + RANGE_SIZE - 1, row_max)
        all_ranges.append((curr, end))
        curr += RANGE_SIZE
            
    # Filter list
    pending = [r for r in all_ranges if r not in completed]
    
    return pending, len(completed), len(all_ranges)

def mark_range_done(table_name, rng):
    """Mark a specific range as complete in Postgres."""
    start, end = rng
    pconn, pcur = pg_conn()
    try:
        pcur.execute("""
            INSERT INTO etl_checkpoints (table_name, start_rowid, end_rowid, status)
            VALUES (%s, %s, %s, 'done')
            ON CONFLICT (table_name, start_rowid, end_rowid) 
            DO UPDATE SET status = 'done', updated_at = now();
        """, (table_name, start, end))
        pconn.commit()
    except Exception as e:
        print(f"⚠️ Failed to save checkpoint: {e}")
    finally:
        pconn.close()

# =======================
# INDEX MANAGEMENT
# =======================
def drop_heavy_indexes():
    print("🗑️  Dropping heavy indexes for faster inserts...")
    pconn, pcur = pg_conn()
    
    indexes_to_drop = [
        "idx_tracks_search",
        "idx_tracks_title_trgm", 
        "idx_tracks_artists_trgm",
        "idx_tracks_popularity",
    ]
    
    for idx in indexes_to_drop:
        try:
            pcur.execute(f"DROP INDEX IF EXISTS {idx}")
            print(f"   Dropped: {idx}")
        except Exception as e:
            print(f"   Skip {idx}: {e}")
    
    pconn.commit()
    pconn.close()
    print("✅ Indexes dropped\n")

def recreate_indexes():
    print("\n🔧 Recreating indexes (this may take a few minutes)...")
    pconn, pcur = pg_conn()
    
    indexes = [
        ("idx_tracks_popularity", "CREATE INDEX idx_tracks_popularity ON tracks (popularity DESC)"),
        ("idx_tracks_title_trgm", "CREATE INDEX idx_tracks_title_trgm ON tracks USING GIN (title gin_trgm_ops)"),
        ("idx_tracks_artists_trgm", "CREATE INDEX idx_tracks_artists_trgm ON tracks USING GIN (artists gin_trgm_ops)"),
        ("idx_tracks_search", "CREATE INDEX idx_tracks_search ON tracks USING GIN (to_tsvector('simple', title || ' ' || artists || ' ' || genres))"),
    ]
    
    for name, sql in tqdm(indexes, desc="📇 Indexes", unit="idx"):
        try:
            start = time.time()
            pcur.execute(sql)
            pconn.commit()
            elapsed = time.time() - start
            tqdm.write(f"   ✅ {name} ({elapsed:.1f}s)")
        except Exception as e:
            tqdm.write(f"   ❌ {name}: {e}")
            pconn.rollback()
    
    pconn.close()
    print("✅ All indexes recreated\n")

# =======================
# ETL PROCESSORS
# =======================
def process_tracks_range(rng):
    start, end = rng
    sconn, scur = sqlite_conn()
    pconn, pcur = pg_conn()
    processed = 0
    
    try:
        # CTE Optimization: Filter tracks FIRST, then join. This prevents full table scans.
        QUERY = """
        WITH batch_tracks AS (
            SELECT rowid, id, name, duration_ms, preview_url, popularity, album_rowid
            FROM tracks
            WHERE rowid BETWEEN ? AND ?
        )
        SELECT
            t.id AS track_id,
            t.name AS title,
            REPLACE(GROUP_CONCAT(DISTINCT a.name), ',', ', ') AS artists,
            REPLACE(GROUP_CONCAT(DISTINCT g.genre), ',', ', ') AS genres,
            t.duration_ms,
            t.preview_url,
            t.popularity,
            (
                SELECT url FROM album_images 
                WHERE album_rowid = t.album_rowid AND width = 64 
                LIMIT 1
            ) AS image_small,
            (
                SELECT url FROM album_images 
                WHERE album_rowid = t.album_rowid AND width >= 500 
                ORDER BY width ASC LIMIT 1
            ) AS image_large
        FROM batch_tracks t
        LEFT JOIN track_artists ta ON ta.track_rowid = t.rowid
        LEFT JOIN artists a ON a.rowid = ta.artist_rowid
        LEFT JOIN artist_genres g ON g.artist_rowid = a.rowid
        GROUP BY t.id
        """

        INSERT_SQL = """
        INSERT INTO tracks (
            track_id, title, artists, genres, duration_ms, 
            image_small, image_large, preview_url, popularity
        ) VALUES %s
        ON CONFLICT (track_id) DO NOTHING
        """

        scur.execute(QUERY, (start, end))
        
        rows = []
        for r in scur:
            rows.append((
                r["track_id"], r["title"], r["artists"] or "", r["genres"] or "",
                r["duration_ms"], r["image_small"], r["image_large"],
                r["preview_url"], r["popularity"]
            ))

            if len(rows) >= BATCH_SIZE:
                execute_values(pcur, INSERT_SQL, rows)
                pconn.commit()
                processed += len(rows)
                rows.clear()

        if rows:
            execute_values(pcur, INSERT_SQL, rows)
            pconn.commit()
            processed += len(rows)
            
        # Success! Mark checkpoint
        mark_range_done("tracks", rng)

    except Exception as e:
        tqdm.write(f"❌ Tracks range {start}-{end}: {e}")
        pconn.rollback()
    finally:
        sconn.close(); pconn.close()
    
    return processed

def process_artists_range(rng):
    start, end = rng
    sconn, scur = sqlite_conn()
    pconn, pcur = pg_conn()
    processed = 0
    try:
        QUERY = """
        SELECT
            a.id AS artist_id, a.name,
            (SELECT url FROM artist_images WHERE artist_rowid = a.rowid AND width = 64 LIMIT 1) AS image_small,
            (SELECT url FROM artist_images WHERE artist_rowid = a.rowid AND width >= 500 ORDER BY width ASC LIMIT 1) AS image_large
        FROM artists a WHERE a.rowid BETWEEN ? AND ?
        """
        INSERT_SQL = """
        INSERT INTO artists (artist_id, name, image_small, image_large) VALUES %s
        ON CONFLICT (artist_id) DO NOTHING
        """
        scur.execute(QUERY, (start, end))
        rows = []
        for r in scur:
            rows.append((r["artist_id"], r["name"], r["image_small"], r["image_large"]))
            if len(rows) >= BATCH_SIZE:
                execute_values(pcur, INSERT_SQL, rows); pconn.commit(); processed += len(rows); rows.clear()
        if rows: execute_values(pcur, INSERT_SQL, rows); pconn.commit(); processed += len(rows)
        
        mark_range_done("artists", rng)
    except Exception as e:
        tqdm.write(f"❌ Artists range {start}-{end}: {e}"); pconn.rollback()
    finally:
        sconn.close(); pconn.close()
    return processed

def process_track_artists_range(rng):
    start, end = rng
    sconn, scur = sqlite_conn()
    pconn, pcur = pg_conn()
    processed = 0
    try:
        QUERY = """
        SELECT t.id, a.id 
        FROM track_artists ta
        JOIN tracks t ON t.rowid = ta.track_rowid
        JOIN artists a ON a.rowid = ta.artist_rowid
        WHERE ta.rowid BETWEEN ? AND ?
        """
        INSERT_SQL = "INSERT INTO track_artists (track_id, artist_id) VALUES %s ON CONFLICT DO NOTHING"
        
        scur.execute(QUERY, (start, end))
        rows = []
        for r in scur:
            rows.append((r[0], r[1]))
            if len(rows) >= BATCH_SIZE:
                execute_values(pcur, INSERT_SQL, rows); pconn.commit(); processed += len(rows); rows.clear()
        if rows: execute_values(pcur, INSERT_SQL, rows); pconn.commit(); processed += len(rows)
        
        mark_range_done("track_artists", rng)
    except Exception as e:
        tqdm.write(f"❌ Relation range {start}-{end}: {e}"); pconn.rollback()
    finally:
        sconn.close(); pconn.close()
    return processed

# =======================
# CORE RUNNER
# =======================
def run_etl_for_table(table_name, process_func, count_query, range_query, start_at=None):
    print(f"\n{'='*60}")
    print(f"📀 Starting ETL for: {table_name}")
    print(f"{'='*60}\n")
    
    # 1. Get Metadata
    sconn, scur = sqlite_conn()
    scur.execute(count_query); total_count = scur.fetchone()[0]
    scur.execute(range_query); row_min, row_max = scur.fetchone()
    sconn.close()
    
    if row_min is None or total_count == 0:
        print(f"❌ No {table_name} found in SQLite.")
        return 0
        
    # Override logic: Jump ahead to start_at if provided
    if start_at is not None and start_at > row_min:
        print(f"⏩ Overriding start rowid to: {start_at:,}")
        row_min = start_at
    
    # 2. Get Pending Ranges
    pending_ranges, done_count, total_chunks = get_pending_ranges(table_name, row_min, row_max)
    
    if not pending_ranges:
        print(f"✅ {table_name} - All valid chunks already completed! Skipping.")
        return 0
    
    # Calculate estimated rows remaining for the progress bar
    # (Adjust estimation if filtering heavily)
    estimated_remaining = int(total_count * (len(pending_ranges) / max(1, total_chunks)))
    
    print(f"📊 Total DB Rows: {total_count:,}")
    if start_at:
        print(f"🎯 Starting from: {start_at:,} (Skipping ~{start_at/total_count:.1%} of DB)")
    print(f"📦 Progress Cache: {done_count} chunks already done")
    print(f"🎯 Pending Work: {len(pending_ranges)} chunks")
    print(f"👷 Workers: {WORKERS} | Batch: {BATCH_SIZE:,}")
    print("-" * 60)
    
    total_processed = 0
    start_time = time.time()
    
    with ThreadPoolExecutor(max_workers=WORKERS) as executor:
        futures = {executor.submit(process_func, rng): rng for rng in pending_ranges}
        
        with tqdm(
            total=len(pending_ranges) * RANGE_SIZE, # Approximate real work
            desc=f"🚀 {table_name}",
            unit="rows",
            colour="green",
            dynamic_ncols=True,
            smoothing=0.05
        ) as pbar:
            for future in as_completed(futures):
                try:
                    processed = future.result()
                    total_processed += processed
                    pbar.update(processed)
                except Exception as e:
                    rng = futures[future]
                    tqdm.write(f"❌ Range {rng} crashed: {e}")

    elapsed = time.time() - start_time
    rate = total_processed / elapsed if elapsed > 0 else 0
    print(f"\n✅ {table_name} batch done! ({elapsed/60:.1f}m)")
    return total_processed

def run_bulk_etl(skip_artists=False, skip_tracks=False, skip_relations=False, start_at=None):
    print("\n🚀 INITIALIZING ETL PIPELINE")
    init_checkpoints()
    drop_heavy_indexes()
    
    # Logic: Apply start_at to the FIRST non-skipped table, then consume it.
    current_start_at = start_at

    if not skip_tracks:
        run_etl_for_table(
            "tracks", process_tracks_range,
            "SELECT COUNT(*) FROM tracks", "SELECT MIN(rowid), MAX(rowid) FROM tracks",
            start_at=current_start_at
        )
        current_start_at = None # Consumed for subequent tables
    
    if not skip_artists:
        run_etl_for_table(
            "artists", process_artists_range,
            "SELECT COUNT(*) FROM artists", "SELECT MIN(rowid), MAX(rowid) FROM artists",
            start_at=current_start_at
        )
        current_start_at = None
    
    if not skip_relations:
        run_etl_for_table(
            "track_artists", process_track_artists_range,
            "SELECT COUNT(*) FROM track_artists", "SELECT MIN(rowid), MAX(rowid) FROM track_artists",
            start_at=current_start_at
        )
    
    recreate_indexes()
    print("\n🎉 ETL PIPELINE COMPLETE!")

# =======================
# MAIN
# =======================
def main():
    parser = argparse.ArgumentParser(description="Spotify ETL: Automated Resume")
    parser.add_argument("--bulk", action="store_true", help="Run optimized ETL (auto-resumes)")
    parser.add_argument("--tracks-only", action="store_true", help="Run only tracks (auto-resumes)")
    parser.add_argument("--skip-artists", action="store_true", help="Skip artists table")
    parser.add_argument("--skip-tracks", action="store_true", help="Skip tracks table")
    parser.add_argument("--skip-relations", action="store_true", help="Skip relations table")
    parser.add_argument("--recreate-indexes", action="store_true", help="Force index rebuild")
    parser.add_argument("--start-at", type=int, help="Force start from a specific rowid for the first table")
    
    args = parser.parse_args()
    
    # Initialize DB table for checkpoints first
    init_checkpoints()
    
    if args.recreate_indexes:
        recreate_indexes()
    elif args.bulk or args.tracks_only:
        run_bulk_etl(
            skip_artists=args.skip_artists or args.tracks_only,
            skip_tracks=args.skip_tracks,
            skip_relations=args.skip_relations or args.tracks_only,
            start_at=args.start_at
        )
    else:
        print("Usage: python main.py --bulk --start-at 100000000")
        print("       (The script will automatically resume from where it left off, or start at the provided rowid)")

if __name__ == "__main__":
    main()