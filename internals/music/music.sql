
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- TRACKS
CREATE TABLE IF NOT EXISTS tracks (
  track_id     TEXT PRIMARY KEY,

  title        TEXT NOT NULL,
  artists      TEXT NOT NULL,
  genres       TEXT NOT NULL,

  duration_ms  INT NOT NULL CHECK (duration_ms >= 0),

  image_small  TEXT,
  image_large  TEXT,
  preview_url  TEXT,
  video_id     TEXT,

  popularity   INT NOT NULL DEFAULT 0,

  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Full text search index
CREATE INDEX IF NOT EXISTS idx_tracks_search
ON tracks
USING GIN (
  to_tsvector('simple', title || ' ' || artists || ' ' || genres)
);

-- Popularity sort index
CREATE INDEX IF NOT EXISTS idx_tracks_popularity
ON tracks (popularity DESC);

-- Trigram indexes for fuzzy matching
CREATE INDEX IF NOT EXISTS idx_tracks_title_trgm
ON tracks USING GIN (title gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_tracks_artists_trgm
ON tracks USING GIN (artists gin_trgm_ops);


-- TRACK STATISTICS
CREATE TABLE IF NOT EXISTS track_stats (
  track_id   TEXT PRIMARY KEY REFERENCES tracks(track_id) ON DELETE CASCADE,

  play_count BIGINT NOT NULL DEFAULT 0,
  listeners  BIGINT NOT NULL DEFAULT 0
);


-- LISTENING HISTORY
CREATE TABLE IF NOT EXISTS history_tracks (
  id        BIGSERIAL PRIMARY KEY,

  user_id   TEXT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  track_id  TEXT      NOT NULL REFERENCES tracks(track_id) ON DELETE CASCADE,

  played_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================
-- STATS TRIGGERS
-- ============================

CREATE OR REPLACE FUNCTION trg_update_track_play_count()
RETURNS trigger AS $$
BEGIN
  INSERT INTO track_stats (track_id, play_count)
  VALUES (NEW.track_id, 1)
  ON CONFLICT (track_id)
  DO UPDATE SET
    play_count = track_stats.play_count + 1;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_history_play_increment
AFTER INSERT ON history_tracks
FOR EACH ROW
EXECUTE FUNCTION trg_update_track_play_count();

-- Fast lookup for "My History"
CREATE INDEX IF NOT EXISTS idx_history_user_time
ON history_tracks (user_id, played_at DESC);

-- Analytics lookup "Who listened to this track"
CREATE INDEX IF NOT EXISTS idx_history_track
ON history_tracks (track_id);


-- USER INTERACTIONS
DO $$ BEGIN
    CREATE TYPE track_interaction_type AS ENUM ('liked', 'saved', 'hidden');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS user_track (
  user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  track_id         TEXT NOT NULL REFERENCES tracks(track_id) ON DELETE CASCADE,

  interaction_type track_interaction_type NOT NULL,
  interacted_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

  PRIMARY KEY (user_id, track_id, interaction_type)
);

CREATE INDEX IF NOT EXISTS idx_user_track_track
ON user_track (track_id);

CREATE INDEX IF NOT EXISTS idx_user_track_user
ON user_track (user_id, interacted_at DESC);

CREATE TABLE artists (
  artist_id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  image_small TEXT,
  image_large TEXT
);

CREATE TABLE track_artists (
  track_id  TEXT NOT NULL REFERENCES tracks(track_id),
  artist_id TEXT NOT NULL REFERENCES artists(artist_id),
  PRIMARY KEY (track_id, artist_id)
);

CREATE INDEX idx_track_artists_artist
ON track_artists (artist_id);
