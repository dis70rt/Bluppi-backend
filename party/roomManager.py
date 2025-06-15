import sys, os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from synqit_db import SynqItDB
from redisManager import RedisManager
import uuid
import time
import hashlib
import logging

class RoomManager():
    def __init__(self):
        self.table_name = "rooms"
        self.redis_manager = RedisManager()

    def create_room(self, name, host_user_id, description=None, visibility="PUBLIC", invite_only=False):
        with SynqItDB() as db:
            room_id = str(uuid.uuid4())
            room_code = self.generate_room_code()
            
            # Create room in database
            db.cursor.execute(
                f"""
                INSERT INTO {self.table_name} (id, name, host_user_id, room_code, description, visibility, invite_only)
                VALUES (%s, %s, %s, %s, %s, %s, %s)
                """,
                (room_id, name, host_user_id, room_code, description, visibility, invite_only)
            )
            db.connection.commit()
            
            # Create Redis room channel
            self.redis_manager.create_room_channel(room_id)
            
            # Set host as connected
            self.redis_manager.set_host_connected(room_id, host_user_id)
            
            # Add host as first member
            self.redis_manager.add_member_to_room(room_id, host_user_id)
            
            # Publish room creation event
            self.redis_manager.publish_room_update(room_id, {
                'type': 'room_created',
                'room_id': room_id,
                'host_user_id': host_user_id,
                'timestamp': time.time()
            })
            
            logging.info(f"Room {room_id} created and channel activated")
            return room_id
    
    def join_room(self, room_id: str, user_id: str) -> bool:
        """Join user to room"""
        try:
            # Check if room is active
            if not self.redis_manager.is_room_active(room_id):
                return False
            
            # Add user to room members
            new_count = self.redis_manager.add_member_to_room(room_id, user_id)
            
            # Publish member join event (incremental update)
            self.redis_manager.publish_room_update(room_id, {
                'type': 'member_join',
                'user_id': user_id,
                'member_count': new_count,
                'timestamp': time.time()
            })
            
            return True
        except Exception as e:
            logging.error(f"Failed to join room {room_id}: {e}")
            return False
    
    def leave_room(self, room_id: str, user_id: str) -> bool:
        """Remove user from room"""
        try:
            # Check if user is host
            host_info = self.redis_manager.redis_client.hgetall(f'room:{room_id}:host')
            is_host = host_info.get('user_id') == user_id
            
            if is_host:
                # Host is leaving, start timeout
                self.redis_manager.set_host_disconnected(room_id)
                
                # Publish host disconnect event
                self.redis_manager.publish_room_update(room_id, {
                    'type': 'host_disconnected',
                    'user_id': user_id,
                    'timeout_minutes': 3,
                    'timestamp': time.time()
                })
            else:
                # Regular member leaving
                new_count = self.redis_manager.remove_member_from_room(room_id, user_id)
                
                # Publish member leave event (incremental update)
                self.redis_manager.publish_room_update(room_id, {
                    'type': 'member_leave',
                    'user_id': user_id,
                    'member_count': new_count,
                    'timestamp': time.time()
                })
            
            return True
        except Exception as e:
            logging.error(f"Failed to leave room {room_id}: {e}")
            return False
    
    def update_playback_state(self, room_id: str, track_id: str = None, position_ms: int = None, status: str = None):
        """Update room playback state and broadcast"""
        try:
            # Update Redis playback state
            updates = {'updated_at': time.time()}
            if track_id is not None:
                updates['track_id'] = track_id
            if position_ms is not None:
                updates['position_ms'] = position_ms
            if status is not None:
                updates['status'] = status
            
            self.redis_manager.redis_client.hmset(f'room:{room_id}:playback', updates)
            
            # Publish playback update (only changes)
            self.redis_manager.publish_room_update(room_id, {
                'type': 'playback_update',
                'changes': updates,
                'timestamp': time.time()
            })
            
        except Exception as e:
            logging.error(f"Failed to update playback state for room {room_id}: {e}")
    
    def generate_room_code(self, length=6):
        timestamp = str(time.time_ns())
        hash_digest = hashlib.sha256(timestamp.encode()).hexdigest().upper()
        clean_chars = ''.join(c for c in hash_digest if c not in '0O1I')
        return clean_chars[:length]