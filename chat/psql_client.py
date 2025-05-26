import psycopg2
import psycopg2.extras
from uuid import UUID
from datetime import datetime
from typing import Optional
import os
import logging
from dotenv import load_dotenv

load_dotenv(override=True)

logging.basicConfig(level=logging.INFO)
log = logging.getLogger(__name__)

HOST = os.environ.get("DB_HOST", "localhost")
PORT = os.environ.get("DB_PORT", 5432)
DB_NAME = os.environ.get("CHAT_DB_NAME")
USER = os.environ.get("DB_USER")
PASSWORD = os.environ.get("DB_PASSWORD")


class PostgresClient:
    def __init__(self):
        try:
            self.conn = psycopg2.connect(
                host=HOST,
                port=PORT,
                dbname=DB_NAME,
                user=USER,
                password=PASSWORD
            )

            psycopg2.extras.register_uuid()
            self.conn.autocommit = True
            log.info("Connected to PostgreSQL database")
        except Exception as e:
            log.error(f"Failed to connect to PostgreSQL: {e}")
            raise

    def close(self):
        if self.conn:
            self.conn.close()

    def insert_conversations(self, conversation_id: UUID, conversation_name: str, is_group: bool = False):
        try:
            with self.conn.cursor() as cursor:
                cursor.execute(
                    "INSERT INTO conversations (conversation_id, conversation_name, is_group) VALUES (%s, %s, %s)",
                    (conversation_id, conversation_name, is_group)
                )
            log.info(f"Inserted conversation {conversation_name} with ID {conversation_id}")
        except Exception as e:
            log.error(f"Failed to insert conversation: {e}")
            raise
    
    def insert_conversation_participants(self, conversation_id: UUID, user_id: UUID):
        try:
            with self.conn.cursor() as cursor:
                cursor.execute(
                    "INSERT INTO conversation_participants (conversation_id, user_id) VALUES (%s, %s)",
                    (conversation_id, user_id)
                )
            log.info(f"Inserted participant {user_id} for conversation {conversation_id}")
        except Exception as e:
            log.error(f"Failed to insert conversation participant: {e}")
            raise
            
    def select_user_conversations(self, user_id: UUID):
        try:
            with self.conn.cursor(cursor_factory=psycopg2.extras.DictCursor) as cursor:
                cursor.execute(
                    "SELECT conversation_id FROM conversation_participants WHERE user_id = %s",
                    (user_id,)
                )
                return cursor.fetchall()
        except Exception as e:
            log.error(f"Failed to select user conversations: {e}")
            raise
            
    def select_conversation_participants(self, conversation_id: UUID):
        try:
            with self.conn.cursor(cursor_factory=psycopg2.extras.DictCursor) as cursor:
                cursor.execute(
                    "SELECT user_id FROM conversation_participants WHERE conversation_id = %s",
                    (conversation_id,)
                )
                return cursor.fetchall()
        except Exception as e:
            log.error(f"Failed to select conversation participants: {e}")
            raise
            
    def select_messages(self, conversation_id: UUID, before_id: Optional[UUID] = None, limit: int = 20):
        try:
            with self.conn.cursor(cursor_factory=psycopg2.extras.DictCursor) as cursor:
                if before_id:
                    cursor.execute(
                        "SELECT created_at FROM messages WHERE message_id = %s",
                        (before_id,)
                    )
                    ref_msg = cursor.fetchone()
                    if ref_msg:
                        cursor.execute(
                            """SELECT message_id, sender_id, message_text, created_at 
                               FROM messages WHERE conversation_id = %s AND created_at < %s 
                               ORDER BY created_at DESC LIMIT %s""",
                            (conversation_id, ref_msg["created_at"], limit)
                        )
                        return cursor.fetchall()
                
                cursor.execute(
                    """SELECT message_id, sender_id, message_text, created_at 
                       FROM messages WHERE conversation_id = %s
                       ORDER BY created_at DESC LIMIT %s""",
                    (conversation_id, limit)
                )
                return cursor.fetchall()
        except Exception as e:
            log.error(f"Failed to select messages: {e}")
            raise
            
    def select_messages_since(self, conversation_id: UUID, since_timestamp: datetime):
        try:
            with self.conn.cursor(cursor_factory=psycopg2.extras.DictCursor) as cursor:
                cursor.execute(
                    """SELECT message_id, sender_id, message_text, created_at 
                       FROM messages WHERE conversation_id = %s AND created_at > %s 
                       ORDER BY created_at ASC""",
                    (conversation_id, since_timestamp)
                )
                return cursor.fetchall()
        except Exception as e:
            log.error(f"Failed to select messages since timestamp: {e}")
            raise
    
    def select_message_info(self, message_id: UUID):
        try:
            with self.conn.cursor(cursor_factory=psycopg2.extras.DictCursor) as cursor:
                cursor.execute(
                    "SELECT sender_id, conversation_id FROM messages WHERE message_id = %s",
                    (message_id,)
                )
                return cursor.fetchone()
        except Exception as e:
            log.error(f"Failed to select message info: {e}")
            raise
            
    def insert_message(self, message_id: UUID, conversation_id: UUID, sender_id: UUID, 
                      message_text: str, created_at: datetime):
        try:
            with self.conn.cursor() as cursor:
                cursor.execute(
                    """INSERT INTO messages (message_id, conversation_id, sender_id, message_text, created_at) 
                       VALUES (%s, %s, %s, %s, %s)""",
                    (message_id, conversation_id, sender_id, message_text, created_at)
                )
            log.info(f"Inserted message {message_id} in conversation {conversation_id}")
        except Exception as e:
            log.error(f"Failed to insert message: {e}")
            raise
            
    def insert_message_status(self, message_id: UUID, user_id: UUID, status: str):
        try:
            with self.conn.cursor() as cursor:
                cursor.execute(
                    "INSERT INTO message_status (message_id, user_id, status) VALUES (%s, %s, %s)",
                    (message_id, user_id, status)
                )
            log.info(f"Inserted message status for message {message_id} and user {user_id}")
        except Exception as e:
            log.error(f"Failed to insert message status: {e}")
            raise
            
    def update_message_status(self, message_id: UUID, user_id: UUID, status: str, updated_at: datetime):
        try:
            with self.conn.cursor() as cursor:
                cursor.execute(
                    "UPDATE message_status SET status = %s, updated_at = %s WHERE message_id = %s AND user_id = %s",
                    (status, updated_at, message_id, user_id)
                )
            log.info(f"Updated message status for message {message_id} and user {user_id}")
        except Exception as e:
            log.error(f"Failed to update message status: {e}")
            raise