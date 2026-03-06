-- Active: 1770643623337@@127.0.0.1@5432@bluppi_music

CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS device_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    fcm_token TEXT NOT NULL UNIQUE, 
    device_type VARCHAR(50) NOT NULL CHECK (device_type IN ('android', 'ios', 'web')),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CHECK (char_length(fcm_token) > 0)
);

-- Fast lookup for a specific user's devices
CREATE INDEX idx_device_tokens_user_id ON device_tokens(user_id);

-- Partial index: Lightning fast lookup for ONLY active devices when sending pushes
CREATE INDEX idx_device_tokens_active ON device_tokens(user_id) WHERE is_active = TRUE;

CREATE TABLE IF NOT EXISTS notification_preferences (
    user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    
    push_notifications_enabled BOOLEAN DEFAULT TRUE,
    email_notifications_enabled BOOLEAN DEFAULT FALSE,
    
    party_invites_enabled BOOLEAN DEFAULT TRUE,
    new_followers_enabled BOOLEAN DEFAULT TRUE,
    follow_request_enabled BOOLEAN DEFAULT TRUE,
    follower_listening_enabled BOOLEAN DEFAULT TRUE,

    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER set_notification_preferences_updated_at
BEFORE UPDATE ON notification_preferences
FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();

CREATE TABLE IF NOT EXISTS notification_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    type VARCHAR(50) NOT NULL, -- e.g., 'PARTY_INVITE', 'NEW_FOLLOWER'
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    
    action_data JSONB, 
    
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Fast pagination when opening the notification inbox
CREATE INDEX idx_notification_history_user_id ON notification_history(user_id, created_at DESC);

-- Partial index: Lightning fast count query for the unread "red dot" UI
CREATE INDEX idx_notification_history_unread ON notification_history(user_id) WHERE is_read = FALSE;


-- Function to initialize preferences automatically
CREATE OR REPLACE FUNCTION trg_create_default_notification_prefs()
RETURNS trigger AS $$
BEGIN
  INSERT INTO notification_preferences (user_id) 
  VALUES (NEW.id);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_default_prefs
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION trg_create_default_notification_prefs();