CREATE TYPE room_status AS ENUM ('ACTIVE', 'INACTIVE', 'CLOSED');
CREATE TYPE room_visibility AS ENUM ('PUBLIC', 'PRIVATE');
CREATE TYPE room_member_role AS ENUM ('HOST', 'MODERATOR', 'PARTICIPANT');
CREATE TYPE playback_status AS ENUM ('PLAYING', 'PAUSED');
CREATE TYPE event_type AS ENUM ('PLAY', 'PAUSE', 'SEEK', 'SKIP');

CREATE TABLE rooms (
    id TEXT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status room_status DEFAULT 'ACTIVE',
    visibility room_visibility DEFAULT 'PUBLIC',
    invite_only BOOLEAN DEFAULT FALSE,
    host_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    current_track_id INT REFERENCES tracks(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE room_members (
    id TEXT PRIMARY KEY,
    room_id TEXT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role room_member_role DEFAULT 'PARTICIPANT',
    invited_by TEXT REFERENCES users(id) ON DELETE SET NULL,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(room_id, user_id, joined_at)
);

CREATE TABLE room_queue (
    id TEXT PRIMARY KEY,
    room_id TEXT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    track_id INT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    added_by TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(room_id, position)
);

CREATE TABLE playback_state (
    room_id TEXT PRIMARY KEY REFERENCES rooms(id) ON DELETE CASCADE,
    current_track_id INT REFERENCES tracks(id) ON DELETE SET NULL,
    position_ms INTEGER DEFAULT 0 CHECK (position_ms >= 0),
    status playback_status DEFAULT 'PAUSED',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE playback_event_log (
    event_id BIGSERIAL PRIMARY KEY,
    room_id TEXT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type event_type NOT NULL,
    event_payload JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX idx_rooms_status ON rooms(status);
CREATE INDEX idx_rooms_visibility ON rooms(visibility);
CREATE INDEX idx_rooms_host_user_id ON rooms(host_user_id);
CREATE INDEX idx_rooms_status_visibility ON rooms(status, visibility);
CREATE INDEX idx_rooms_created_at ON rooms(created_at DESC);

CREATE INDEX idx_room_members_room_id ON room_members(room_id);
CREATE INDEX idx_room_members_user_id ON room_members(user_id);
CREATE INDEX idx_room_members_role ON room_members(role);
CREATE INDEX idx_room_members_active ON room_members(room_id, user_id) WHERE left_at IS NULL;
CREATE INDEX idx_room_members_joined_at ON room_members(joined_at DESC);

CREATE INDEX idx_room_queue_room_id ON room_queue(room_id);
CREATE INDEX idx_room_queue_position ON room_queue(room_id, position);
CREATE INDEX idx_room_queue_added_by ON room_queue(added_by);
CREATE INDEX idx_room_queue_added_at ON room_queue(added_at DESC);

CREATE INDEX idx_playback_event_log_room_id ON playback_event_log(room_id);
CREATE INDEX idx_playback_event_log_user_id ON playback_event_log(user_id);
CREATE INDEX idx_playback_event_log_timestamp ON playback_event_log(timestamp DESC);
CREATE INDEX idx_playback_event_log_event_type ON playback_event_log(event_type);
CREATE INDEX idx_playback_event_log_room_timestamp ON playback_event_log(room_id, timestamp DESC);

-- Triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_rooms_updated_at BEFORE UPDATE ON rooms FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_playback_state_updated_at BEFORE UPDATE ON playback_state FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to get active room members
CREATE OR REPLACE FUNCTION get_active_room_members(p_room_id TEXT)
RETURNS TABLE (
    user_id TEXT,
    username TEXT,
    name TEXT,
    avatar_url TEXT,
    role room_member_role,
    joined_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        u.id,
        u.username,
        u.name,
        u.profile_pic,
        rm.role,
        rm.joined_at
    FROM room_members rm
    JOIN users u ON rm.user_id = u.id
    WHERE rm.room_id = p_room_id 
    AND rm.left_at IS NULL
    ORDER BY rm.joined_at;
END;
$$ LANGUAGE plpgsql;

-- Function to get room queue in order
CREATE OR REPLACE FUNCTION get_room_queue(p_room_id TEXT)
RETURNS TABLE (
    track_index INTEGER,
    track_id INT,
    title VARCHAR(255),
    artist VARCHAR(255),
    image_url VARCHAR(255),
    duration INT,
    added_by_username TEXT,
    added_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        rq.position,
        t.id,
        t.title,
        t.artist,
        t.image_url,
        t.duration,
        u.username,
        rq.added_at
    FROM room_queue rq
    JOIN tracks t ON rq.track_id = t.id
    JOIN users u ON rq.added_by = u.id
    WHERE rq.room_id = p_room_id
    ORDER BY rq.position;
END;
$$ LANGUAGE plpgsql;

ALTER TABLE room_members ADD CONSTRAINT check_left_after_joined 
    CHECK (left_at IS NULL OR left_at >= joined_at);
ALTER TABLE room_queue ADD CONSTRAINT check_positive_position 
    CHECK (position > 0);

CREATE INDEX idx_rooms_active ON rooms(id, name) WHERE status = 'ACTIVE';
CREATE INDEX idx_rooms_public_active ON rooms(visibility, status, created_at DESC) 
    WHERE visibility = 'PUBLIC' AND status = 'ACTIVE';

ALTER TABLE rooms ADD room_code VARCHAR(7) UNIQUE NOT NULL;

CREATE OR REPLACE FUNCTION status_change_of_room()
RETURN TRIGGER AS $$
