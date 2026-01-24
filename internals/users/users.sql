-- =========================
-- USERS TABLE
-- =========================
CREATE TABLE users (
  id               TEXT        PRIMARY KEY,           -- auth-provider user id
  email            TEXT        NOT NULL,
  username         TEXT        NOT NULL,
  name             TEXT        NOT NULL,
  bio              TEXT,
  country          TEXT,
  phone            TEXT,
  profile_pic      TEXT,
  favorite_genres  TEXT[]      NOT NULL DEFAULT '{}',
  follower_count   INT         NOT NULL DEFAULT 0,
  following_count  INT         NOT NULL DEFAULT 0,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Case-insensitive uniqueness
CREATE UNIQUE INDEX users_email_ci
  ON users ((LOWER(email)));

CREATE UNIQUE INDEX users_username_ci
  ON users ((LOWER(username)));



-- =========================
-- FOLLOWS (USER ↔ USER)
-- =========================
CREATE TABLE follows (
  follower_id TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  followee_id TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (follower_id, followee_id),
  CHECK (follower_id <> followee_id)
);

-- Read-optimized indexes
CREATE INDEX idx_follows_by_followee
  ON follows (followee_id, created_at DESC);

CREATE INDEX idx_follows_by_follower
  ON follows (follower_id, created_at DESC);



-- =========================
-- FOLLOW COUNTER TRIGGER
-- =========================
CREATE OR REPLACE FUNCTION trg_update_follow_counts()
RETURNS trigger AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    UPDATE users
      SET following_count = following_count + 1
      WHERE id = NEW.follower_id;

    UPDATE users
      SET follower_count = follower_count + 1
      WHERE id = NEW.followee_id;

    RETURN NEW;

  ELSIF TG_OP = 'DELETE' THEN
    UPDATE users
      SET following_count = GREATEST(following_count - 1, 0)
      WHERE id = OLD.follower_id;

    UPDATE users
      SET follower_count = GREATEST(follower_count - 1, 0)
      WHERE id = OLD.followee_id;

    RETURN OLD;
  END IF;

  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_follows_counts
AFTER INSERT OR DELETE ON follows
FOR EACH ROW
EXECUTE FUNCTION trg_update_follow_counts();



-- =========================
-- RECENT SEARCHES
-- =========================
CREATE TABLE recent_searches (
  id          BIGSERIAL    PRIMARY KEY,
  user_id     TEXT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  query       TEXT         NOT NULL,
  searched_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- Full-text search on queries
CREATE INDEX idx_searches_query_fts
  ON recent_searches
  USING GIN (to_tsvector('simple', query));

-- Fast lookup by user
CREATE INDEX idx_searches_by_user
  ON recent_searches (user_id, searched_at DESC);
