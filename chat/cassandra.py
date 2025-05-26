from cassandra.cluster import Cluster, Session, PreparedStatement
from uuid import UUID
from datetime import datetime
from typing import List, Optional, Dict, Any

import os
import logging

from dotenv import load_dotenv

load_dotenv(override=True)

logging.basicConfig(level=logging.INFO)
log = logging.getLogger(__name__)

HOST = os.environ("CASSANDRA_HOST", "localhost")
PORT = os.environ("CASSANDRA_PORT", 9042)
KEY_SPACE = None


class PreparedQuery:
    def __init__(self, session: Session):
        self.insert_conversations: PreparedStatement = session.prepare(
            "INSERT INTO conversations (conversation_id, conversation_name, is_group) VALUES (?, ?, ?)"
        )
        self.insert_conversation_participants: PreparedStatement = session.prepare(
            "INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)"
        )
        self.select_user_conversations: PreparedStatement = session.prepare(
            "SELECT conversation_id FROM conversation_participants WHERE user_id = ?"
        )
        self.select_conversation_participants: PreparedStatement = session.prepare(
            "SELECT user_id FROM conversation_participants WHERE conversation_id = ?"
        )
        self.select_messages: PreparedStatement = session.prepare(
            "SELECT message_id, sender_id, message_text, created_at FROM messages WHERE conversation_id = ?"
        )
        self.select_messages_before: PreparedStatement = session.prepare(
            "SELECT message_id, sender_id, message_text, created_at FROM messages WHERE conversation_id = ? AND created_at < ? ORDER BY created_at DESC LIMIT ?"
        )
        self.select_messages_since: PreparedStatement = session.prepare(
            "SELECT message_id, sender_id, message_text, created_at FROM messages WHERE conversation_id = ? AND created_at > ? ORDER BY created_at ASC"
        )
        self.select_message_timestamp: PreparedStatement = session.prepare(
            "SELECT created_at FROM messages WHERE message_id = ?"
        )
        self.select_message_info: PreparedStatement = session.prepare(
            "SELECT sender_id, conversation_id FROM messages WHERE message_id = ?"
        )
        self.insert_message: PreparedStatement = session.prepare(
            "INSERT INTO messages (message_id, conversation_id, sender_id, message_text, created_at) VALUES (?, ?, ?, ?, ?)"
        )
        self.insert_message_status: PreparedStatement = session.prepare(
            "INSERT INTO message_status (message_id, user_id, status) VALUES (?, ?, ?)"
        )
        self.update_message_status: PreparedStatement = session.prepare(
            "UPDATE message_status SET status = ?, updated_at = ? WHERE message_id = ? AND user_id = ?"
        )


class CassandraClient:
    def __init__(self):
        try:
            self.cluster = Cluster([HOST], port=PORT)
            self.session = self.cluster.connect(KEY_SPACE)
            self.prepared_query = PreparedQuery(self.session)
            log.info("Connected to Cassandra")
        except Exception as e:
            log.error(f"Failed to connect to Cassandra: {e}")
            raise

    def close(self):
        self.cluster.shutdown()

    def insert_conversations(self, convsersation_id: UUID, conversation_name: str, is_group: bool = False):
        try:
            self.session.execute(
                self.prepared_query.insert_conversations.bind(
                    (convsersation_id, conversation_name, is_group)
                )
            )
            log.info(f"Inserted conversation {conversation_name} with ID {convsersation_id}")
        except Exception as e:
            log.error(f"Failed to insert conversation: {e}")
            raise
    
    def insert_conversation_participants(self, conversation_id: UUID, user_id: str):
        try:
            self.session.execute(
                self.prepared_query.insert_conversation_participants.bind(
                    (conversation_id, user_id)
                )
            )
            log.info(f"Inserted participant {user_id} for conversation {conversation_id}")
        except Exception as e:
            log.error(f"Failed to insert conversation participant: {e}")
            raise
            
    def select_user_conversations(self, user_id: UUID):
        try:
            rows = self.session.execute(
                self.prepared_query.select_user_conversations.bind((user_id,))
            )
            return rows
        except Exception as e:
            log.error(f"Failed to select user conversations: {e}")
            raise
            
    def select_conversation_participants(self, conversation_id: UUID):
        try:
            rows = self.session.execute(
                self.prepared_query.select_conversation_participants.bind((conversation_id,))
            )
            return rows
        except Exception as e:
            log.error(f"Failed to select conversation participants: {e}")
            raise
            
    def select_messages(self, conversation_id: UUID, before_id: Optional[UUID] = None, limit: int = 20):
        try:
            if before_id:
                ref_msg = self.session.execute(
                    self.prepared_query.select_message_timestamp.bind((before_id,))
                ).one()
                if ref_msg:
                    rows = self.session.execute(
                        self.prepared_query.select_messages_before.bind(
                            (conversation_id, ref_msg.created_at, limit)
                        )
                    )
                    return rows
            
            rows = self.session.execute(
                self.prepared_query.select_messages.bind((conversation_id,))
            )
            return rows
        except Exception as e:
            log.error(f"Failed to select messages: {e}")
            raise
            
    def select_messages_since(self, conversation_id: UUID, since_timestamp: datetime):
        try:
            rows = self.session.execute(
                self.prepared_query.select_messages_since.bind(
                    (conversation_id, since_timestamp)
                )
            )
            return rows
        except Exception as e:
            log.error(f"Failed to select messages since timestamp: {e}")
            raise
    
    def select_message_info(self, message_id: UUID):
        try:
            row = self.session.execute(
                self.prepared_query.select_message_info.bind((message_id,))
            ).one()
            return row
        except Exception as e:
            log.error(f"Failed to select message info: {e}")
            raise
            
    def insert_message(self, message_id: UUID, conversation_id: UUID, sender_id: UUID, 
                      message_text: str, created_at: datetime):
        try:
            self.session.execute(
                self.prepared_query.insert_message.bind(
                    (message_id, conversation_id, sender_id, message_text, created_at)
                )
            )
            log.info(f"Inserted message {message_id} in conversation {conversation_id}")
        except Exception as e:
            log.error(f"Failed to insert message: {e}")
            raise
            
    def insert_message_status(self, message_id: UUID, user_id: UUID, status: str):
        try:
            self.session.execute(
                self.prepared_query.insert_message_status.bind(
                    (message_id, user_id, status)
                )
            )
            log.info(f"Inserted message status for message {message_id} and user {user_id}")
        except Exception as e:
            log.error(f"Failed to insert message status: {e}")
            raise
            
    def update_message_status(self, message_id: UUID, user_id: UUID, status: str):
        try:
            self.session.execute(
                self.prepared_query.update_message_status.bind(
                    (status, datetime.now(), message_id, user_id)
                )
            )
            log.info(f"Updated message status for message {message_id} and user {user_id}")
        except Exception as e:
            log.error(f"Failed to update message status: {e}")
            raise