-- =========================
-- EXTENSIONS
-- =========================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";



-- =========================
-- TRACKS
-- =========================
CREATE TABLE tracks (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  title        TEXT NOT NULL,
  artist       TEXT NOT NULL,
  album        TEXT,
  duration     INT CHECK (duration >= 0),
  genre        TEXT[] NOT NULL DEFAULT '{}',
  image_url    TEXT,
  preview_url  TEXT,
  video_id     TEXT,

  listeners    INT NOT NULL DEFAULT 0,
  play_count   INT NOT NULL DEFAULT 0,
  popularity   INT NOT NULL DEFAULT 0,

  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Search / filter indexes
CREATE INDEX idx_tracks_title_fts
  ON tracks USING GIN (to_tsvector('simple', title));

CREATE INDEX idx_tracks_artist
  ON tracks (artist);

CREATE INDEX idx_tracks_genre
  ON tracks USING GIN (genre);



-- =========================
-- PLAY HISTORY
-- =========================
CREATE TABLE history_tracks (
  id        BIGSERIAL PRIMARY KEY,
  user_id   TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  track_id  UUID NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
  played_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_history_by_user
  ON history_tracks (user_id, played_at DESC);

CREATE INDEX idx_history_by_track
  ON history_tracks (track_id, played_at DESC);



-- =========================
-- USER ↔ TRACK INTERACTIONS
-- =========================
CREATE TYPE track_interaction_type AS ENUM (
  'liked',
  'last_played',
  'most_played',
  'history'
);

CREATE TABLE user_track (
  user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  track_id         UUID NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
  interaction_type track_interaction_type NOT NULL,
  interacted_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, track_id, interaction_type)
);

CREATE INDEX idx_user_track_by_user
  ON user_track (user_id, interacted_at DESC);

CREATE INDEX idx_user_track_by_track
  ON user_track (track_id);



-- =========================
-- PLAY COUNT + LISTENER TRIGGER
-- =========================
CREATE OR REPLACE FUNCTION trg_update_track_stats()
RETURNS trigger AS $$
BEGIN
  -- increment play count
  UPDATE tracks
    SET play_count = play_count + 1
    WHERE id = NEW.track_id;

  -- increment listeners only on first play by user
  IF NOT EXISTS (
    SELECT 1
    FROM history_tracks
    WHERE user_id = NEW.user_id
      AND track_id = NEW.track_id
      AND id <> NEW.id
  ) THEN
    UPDATE tracks
      SET listeners = listeners + 1
      WHERE id = NEW.track_id;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_track_stats
AFTER INSERT ON history_tracks
FOR EACH ROW
EXECUTE FUNCTION trg_update_track_stats();
