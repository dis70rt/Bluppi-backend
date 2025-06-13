import sys, os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from database import SynqItDB
import uuid
import time
import hashlib

class RoomManager():
    def __init__(self):
        self.table_name = "rooms"

    def create_room(self, name, host_user_id, description=None, visibility="PUBLIC",invite_only=False):
        with SynqItDB() as db:
            room_id = str(uuid.uuid4())
            room_code = self.generate_room_code()
            db.cursor.execute(
                f"""
                INSERT INTO {self.table_name} (id, name, host_user_id, room_code, description, visibility, invite_only)
                VALUES (%s, %s, %s, %s, %s, %s, %s)
                """,
                (room_id, name, host_user_id, room_code, description, visibility, invite_only)
            )
            db.connection.commit()
            return room_id
    
    
    def generate_room_code(self,length=6):
        timestamp = str(time.time_ns())
        hash_digest = hashlib.sha256(timestamp.encode()).hexdigest().upper()
        clean_chars = ''.join(c for c in hash_digest if c not in '0O1I')
        return clean_chars[:length]
    