-- 1. Allow much larger WAL before checkpoints
ALTER SYSTEM SET max_wal_size = '8GB';
ALTER SYSTEM SET min_wal_size = '2GB';

-- 2. Slow down checkpoints, spread IO
ALTER SYSTEM SET checkpoint_timeout = '30min';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;

-- 3. Reduce WAL overhead
ALTER SYSTEM SET wal_compression = on;

-- 4. Disable autovacuum (TEMPORARY)
ALTER SYSTEM SET autovacuum = off;

-- 5. Apply settings
SELECT pg_reload_conf();

-- Reduce fsync pressure during ETL (still crash-safe via WAL)
ALTER SYSTEM SET synchronous_commit = off;
SELECT pg_reload_conf();

DROP INDEX IF EXISTS idx_tracks_search;
DROP INDEX IF EXISTS idx_tracks_title_trgm;
DROP INDEX IF EXISTS idx_tracks_artists_trgm;
