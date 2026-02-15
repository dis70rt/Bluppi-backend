CREATE TABLE IF NOT EXISTS etl_track_checkpoints (
  start_rowid BIGINT NOT NULL,
  end_rowid   BIGINT NOT NULL,
  status      TEXT NOT NULL CHECK (status IN ('pending', 'running', 'done')),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (start_rowid, end_rowid)
);


-- TRUNCATE TABLE
--   tracks,
--   track_stats,
--   history_tracks,
--   user_track,
--   track_artists,
--   artists
-- RESTART IDENTITY
-- CASCADE;
