-- 1. Re-enable autovacuum
ALTER SYSTEM SET autovacuum = on;

-- 2. Restore WAL/checkpoint defaults
ALTER SYSTEM SET max_wal_size = '1GB';
ALTER SYSTEM SET min_wal_size = '80MB';

ALTER SYSTEM SET checkpoint_timeout = '5min';
ALTER SYSTEM SET checkpoint_completion_target = 0.5;

-- 3. Restore durability behavior
ALTER SYSTEM SET synchronous_commit = on;

-- 4. Apply restored settings
SELECT pg_reload_conf();

-- Full-text search
CREATE INDEX idx_tracks_search
ON tracks
USING GIN (
  to_tsvector('simple', title || ' ' || artists || ' ' || genres)
);

-- Trigram fuzzy search
CREATE INDEX idx_tracks_title_trgm
ON tracks USING GIN (title gin_trgm_ops);

CREATE INDEX idx_tracks_artists_trgm
ON tracks USING GIN (artists gin_trgm_ops);
