CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TYPE room_status AS ENUM ('ACTIVE', 'ENDED');
CREATE TYPE room_visibility AS ENUM ('PUBLIC', 'PRIVATE');
CREATE TYPE room_member_role AS ENUM ('HOST', 'LISTENER');

CREATE TABLE rooms (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    code VARCHAR(16) UNIQUE NOT NULL,
    status room_status NOT NULL DEFAULT 'ACTIVE',
    visibility room_visibility NOT NULL DEFAULT 'PUBLIC',
    invite_only BOOLEAN NOT NULL DEFAULT FALSE,
    host_user_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (char_length(code) >= 4)
);

CREATE UNIQUE INDEX uniq_active_room_per_host
ON rooms(host_user_id)
WHERE status = 'ACTIVE';

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_rooms_updated_at
BEFORE UPDATE ON rooms
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE room_members (
    room_id UUID REFERENCES rooms(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    role room_member_role NOT NULL DEFAULT 'LISTENER',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at TIMESTAMPTZ,
    PRIMARY KEY (room_id, user_id)
);

CREATE INDEX idx_rooms_created_at ON rooms(created_at DESC);
CREATE INDEX idx_rooms_name ON rooms(name);
CREATE INDEX idx_room_members_room_id ON room_members(room_id);
CREATE INDEX idx_room_members_active ON room_members(room_id) WHERE left_at IS NULL;

ALTER TABLE room_members
ALTER COLUMN user_id TYPE TEXT
USING user_id::TEXT;

ALTER TABLE rooms
ALTER COLUMN host_user_id TYPE TEXT
USING host_user_id::TEXT;