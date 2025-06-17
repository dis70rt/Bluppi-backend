CREATE TABLE users (
  id               TEXT        PRIMARY KEY,           
  email            TEXT        NOT NULL UNIQUE,
  username         TEXT        NOT NULL UNIQUE,
  name             TEXT        NOT NULL,
  bio              TEXT,
  country          TEXT,
  phone            TEXT,
  profile_pic      TEXT,
  favorite_genres  TEXT[]      NOT NULL DEFAULT '{}', 
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE TABLE follows (
  follower_id  TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  followee_id  TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (follower_id, followee_id)
);


CREATE TABLE history_tracks (
  id         BIGSERIAL    PRIMARY KEY,
  user_id    TEXT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  track_id   UUID         NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
  played_at  TIMESTAMPTZ  NOT NULL
);


CREATE TABLE recent_searches (
  id          BIGSERIAL    PRIMARY KEY,
  user_id     TEXT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  query       TEXT         NOT NULL,
  searched_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TYPE track_interaction_type AS ENUM (
  'liked',
  'last_played',
  'most_played',
);


CREATE TABLE user_track (
  user_id          TEXT                    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  track_id         UUID                    NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
  interaction_type track_interaction_type  NOT NULL,
  interacted_at    TIMESTAMPTZ             NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, track_id, interaction_type)
);


ALTER TABLE users
  ADD COLUMN follower_count  INT NOT NULL DEFAULT 0,
  ADD COLUMN following_count INT NOT NULL DEFAULT 0;


CREATE OR REPLACE FUNCTION trg_update_follow_counts() RETURNS trigger AS $$
BEGIN
  IF (TG_OP = 'INSERT') THEN
    UPDATE users SET following_count = following_count + 1 WHERE id = NEW.follower_id;
    UPDATE users SET follower_count = follower_count + 1 WHERE id = NEW.followee_id;
    RETURN NEW;
  ELSIF (TG_OP = 'DELETE') THEN
    UPDATE users SET following_count = following_count - 1 WHERE id = OLD.follower_id;
    UPDATE users SET follower_count = follower_count - 1 WHERE id = OLD.followee_id;
    RETURN OLD;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_follows_counts
  AFTER INSERT OR DELETE ON follows
  FOR EACH ROW EXECUTE FUNCTION trg_update_follow_counts();


CREATE INDEX idx_follows_by_followee
  ON follows(followee_id, created_at DESC);
CREATE INDEX idx_follows_by_follower
  ON follows(follower_id, created_at DESC);

CREATE INDEX idx_history_by_user
  ON history_tracks(user_id, played_at DESC);

CREATE INDEX idx_searches_query_fts
  ON recent_searches USING GIN (to_tsvector('simple', query));

CREATE UNIQUE INDEX users_username_ci
  ON users ((LOWER(username)));

CREATE UNIQUE INDEX users_email_ci
  ON users ((LOWER(email)));

ALTER TYPE track_interaction_type
  ADD VALUE 'history';

