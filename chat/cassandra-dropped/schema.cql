-- cqlsh 127.0.0.1 9042 -f /home/ethernode/Projects/Bluppi/synqit-py/chat/schema.cql

CREATE KEYSPACE IF NOT EXISTS chat_app
WITH replication = {
  'class' : 'SimpleStrategy',
  'replication_factor' : 1
};

USE chat_app;

CREATE TABLE IF NOT EXISTS conversations (
    conversation_id    UUID PRIMARY KEY,
    conversation_name  TEXT,
    is_group           BOOLEAN,
    created_at         TIMESTAMP
);

CREATE TABLE IF NOT EXISTS conversation_participants (
    conversation_id UUID,
    user_id         UUID,
    joined_at       TIMESTAMP,
    PRIMARY KEY (conversation_id, user_id)
);

CREATE TABLE IF NOT EXISTS messages (
    message_id      UUID PRIMARY KEY,
    conversation_id UUID,
    sender_id       UUID,
    message_text    TEXT,
    created_at      TIMESTAMP
);

CREATE TABLE IF NOT EXISTS message_status (
    message_id  UUID,
    user_id     UUID,
    status      TEXT,
    updated_at  TIMESTAMP,
    PRIMARY KEY (message_id, user_id)
);

CREATE INDEX ON conversation_participants(user_id);
CREATE INDEX ON messages(conversation_id);