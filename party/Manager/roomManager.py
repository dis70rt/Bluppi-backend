import sys, os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from synqit_db import SynqItDB
from redisManager import RedisManager
import uuid
import time
import hashlib
import logging
import json

class RoomManager():
    def __init__(self):
        self.table_name = "rooms"
        self.redis_manager = RedisManager()

    def create_room(self, name, host_user_id, description=None, visibility="PUBLIC", invite_only=False):
        with SynqItDB() as db:
            room_id = str(uuid.uuid4())
            room_code = self.generate_room_code()
            
            try:
                db.cursor.execute(
                    f"""
                    INSERT INTO {self.table_name} (id, name, host_user_id, room_code, description, visibility, invite_only)
                    VALUES (%s, %s, %s, %s, %s, %s, %s)
                    RETURNING id
                    """,
                    (room_id, name, host_user_id, room_code, description, visibility, invite_only)
                )
                room_id = db.cursor.fetchone()[0]
                
                db.cursor.execute(
                    """
                    INSERT INTO room_members (id, room_id, user_id, role)
                    VALUES (%s, %s, %s, 'HOST')
                    """,
                    (str(uuid.uuid4()), room_id, host_user_id)
                )
                
                db.cursor.execute(
                    """
                    INSERT INTO playback_state (room_id, status)
                    VALUES (%s, 'PAUSED')
                    """,
                    (room_id,)
                )
                
                db.connection.commit()
                
                if not self.redis_manager.create_room_channel(room_id):
                    raise Exception("Failed to create Redis room channel")
                
                self.redis_manager.set_host_connected(room_id, host_user_id)                
                self.redis_manager.add_member_to_room(room_id, host_user_id)
                
                self.redis_manager.publish_room_update(room_id, {
                    'type': 'room_created',
                    'room_id': room_id,
                    'host_user_id': host_user_id,
                    'timestamp': time.time()
                })
                
                logging.info(f"Room {room_id} created and channel activated")
                return room_id
                
            except Exception as e:
                logging.error(f"Failed to create room: {e}")
                db.connection.rollback()
                raise e
    
    def join_room(self, room_id: str, user_id: str) -> bool:
        try:
            if not self.redis_manager.is_room_active(room_id):
                logging.warning(f"Room {room_id} is not active")
                return False
            
            with SynqItDB() as db:
                db.cursor.execute(
                    """
                    SELECT id FROM room_members 
                    WHERE room_id = %s AND user_id = %s AND left_at IS NULL
                    """,
                    (room_id, user_id)
                )
                existing = db.cursor.fetchone()
                
                if not existing:
                    db.cursor.execute(
                        """
                        INSERT INTO room_members (id, room_id, user_id, role)
                        VALUES (%s, %s, %s, 'PARTICIPANT')
                        """,
                        (str(uuid.uuid4()), room_id, user_id)
                    )
                    db.connection.commit()
            
            new_count = self.redis_manager.add_member_to_room(room_id, user_id)
            
            self.redis_manager.publish_room_update(room_id, {
                'type': 'member_join',
                'room_id': room_id,
                'user_id': user_id,
                'member_count': new_count,
                'timestamp': time.time()
            })
            
            logging.info(f"User {user_id} joined room {room_id} (total: {new_count})")
            return True
        except Exception as e:
            logging.error(f"Failed to join room {room_id}: {e}")
            return False
    
    def leave_room(self, room_id: str, user_id: str) -> bool:
        try:
            host_info = self.redis_manager.redis_client.hgetall(f'room:{room_id}:host')
            is_host = host_info.get('user_id') == user_id
            
            with SynqItDB() as db:
                db.cursor.execute(
                    """
                    UPDATE room_members
                    SET left_at = CURRENT_TIMESTAMP
                    WHERE room_id = %s AND user_id = %s AND left_at IS NULL
                    """,
                    (room_id, user_id)
                )
                
                if is_host:
                    db.cursor.execute(
                        """
                        UPDATE rooms
                        SET status = 'INACTIVE'
                        WHERE id = %s
                        """,
                        (room_id,)
                    )
                
                db.connection.commit()
            
            if is_host:
                self.redis_manager.set_host_disconnected(room_id)
            
                self.redis_manager.publish_room_update(room_id, {
                    'type': 'host_disconnected',
                    'room_id': room_id,
                    'user_id': user_id,
                    'timeout_minutes': 3,
                    'timestamp': time.time()
                })
                logging.info(f"Host {user_id} disconnected from room {room_id}")
            else:
                new_count = self.redis_manager.remove_member_from_room(room_id, user_id)
                
                self.redis_manager.publish_room_update(room_id, {
                    'type': 'member_leave',
                    'room_id': room_id,
                    'user_id': user_id,
                    'member_count': new_count,
                    'timestamp': time.time()
                })
                logging.info(f"User {user_id} left room {room_id} (remaining: {new_count})")
            
            return True
        except Exception as e:
            logging.error(f"Failed to leave room {room_id}: {e}")
            return False

    def update_playback_state(self, room_id: str, user_id: str, track_id: uuid.UUID = None, position_ms: int = None, status: str = None):
        try:
            updates = {}
            event_type = None
            event_payload = {}
            
            with SynqItDB() as db:
                if track_id is not None:
                    db.cursor.execute(
                        """
                        UPDATE playback_state
                        SET current_track_id = %s, updated_at = CURRENT_TIMESTAMP
                        WHERE room_id = %s
                        """,
                        (track_id, room_id)
                    )
                    updates['track_id'] = str(track_id)
                    event_type = 'SKIP'
                    event_payload['track_id'] = str(track_id)
                    
                if position_ms is not None:
                    db.cursor.execute(
                        """
                        UPDATE playback_state
                        SET position_ms = %s, updated_at = CURRENT_TIMESTAMP
                        WHERE room_id = %s
                        """,
                        (position_ms, room_id)
                    )
                    updates['position_ms'] = position_ms
                    if not event_type:
                        event_type = 'SEEK'
                        event_payload['position_ms'] = position_ms
                    
                if status is not None:
                    db.cursor.execute(
                        """
                        UPDATE playback_state
                        SET status = %s, updated_at = CURRENT_TIMESTAMP
                        WHERE room_id = %s
                        """,
                        (status, room_id)
                    )
                    updates['status'] = status
                    if not event_type:
                        event_type = 'PLAY' if status == 'PLAYING' else 'PAUSE'
                
                if event_type:
                    db.cursor.execute(
                        """
                        INSERT INTO playback_event_log (room_id, user_id, event_type, event_payload)
                        VALUES (%s, %s, %s, %s)
                        """,
                        (room_id, user_id, event_type, json.dumps(event_payload))
                    )
                
                db.connection.commit()
            
            self.redis_manager.update_playback_state(room_id, updates)
            
            logging.info(f"Updated playback state for room {room_id}: {updates}")
            
        except Exception as e:
            logging.error(f"Failed to update playback state for room {room_id}: {e}")
    
    def get_room_info(self, room_id: str) -> dict:
        try:
            with SynqItDB() as db:
                query = f"""
                SELECT r.id, r.name, r.description, r.room_code, r.host_user_id, 
                    r.visibility, r.invite_only, r.created_at, r.updated_at, r.status
                FROM {self.table_name} r
                WHERE r.id = %s
                """
                db.cursor.execute(query, (room_id,))
                room = db.cursor.fetchone()
                
                if not room:
                    return None
                    
                room_dict = {
                    'id': room[0],
                    'name': room[1],
                    'description': room[2] or '',
                    'room_code': room[3],
                    'host_user_id': room[4],
                    'visibility': room[5],
                    'status': room[9],
                    'invite_only': bool(room[6]),
                    'created_at': room[7],
                    'updated_at': room[8]
                }
                
                db.cursor.execute("SELECT * FROM get_active_room_members(%s)", (room_id,))
                members = db.cursor.fetchall()
                room_dict['members'] = []
                
                for member in members:
                    room_dict['members'].append({
                        'user_id': member[0],
                        'username': member[1],
                        'name': member[2],
                        'avatar_url': member[3],
                        'role': member[4],
                        'joined_at': member[5]
                    })
                
                db.cursor.execute(
                    """
                    SELECT current_track_id, position_ms, status, updated_at
                    FROM playback_state
                    WHERE room_id = %s
                    """,
                    (room_id,)
                )
                playback = db.cursor.fetchone()
                if playback:
                    room_dict['playback'] = {
                        'current_track_id': str(playback[0]) if playback[0] else None,
                        'position_ms': playback[1],
                        'status': playback[2],
                        'updated_at': playback[3]
                    }
                
                return room_dict
            
        except Exception as e:
            logging.error(f"Error fetching room info: {e}")
            return None

    def list_active_rooms(self, visibility_filter=None, host_user_id_filter=None, include_private_rooms_if_member=False):
        try:
            with SynqItDB() as db:
                query = f"""
                SELECT r.id, r.name, r.description, r.room_code, r.host_user_id, 
                    r.visibility, r.invite_only, r.created_at, r.updated_at, r.status
                FROM {self.table_name} r
                WHERE r.status = 'ACTIVE'
                """
                
                params = []
                
                if visibility_filter:
                    query += " AND r.visibility = %s"
                    params.append(visibility_filter)
                    
                if host_user_id_filter:
                    query += " AND r.host_user_id = %s"
                    params.append(host_user_id_filter)
                    
                query += " ORDER BY r.created_at DESC"
                    
                db.cursor.execute(query, params)
                rooms = db.cursor.fetchall()
                
                result = []
                for room in rooms:
                    room_dict = {
                        'id': room[0],
                        'name': room[1],
                        'description': room[2] or '',
                        'room_code': room[3],
                        'host_user_id': room[4],
                        'visibility': room[5],
                        'invite_only': bool(room[6]),
                        'created_at': room[7],
                        'updated_at': room[8],
                        'status': room[9]
                    }
                    result.append(room_dict)
                    
                return result
                
        except Exception as e:
            logging.error(f"Error listing active rooms: {e}")
            return []

    def get_room_queue(self, room_id: str) -> list:
        try:
            with SynqItDB() as db:
                db.cursor.execute("SELECT * FROM get_room_queue(%s)", (room_id,))
                queue = db.cursor.fetchall()
                
                queue_list = []
                for item in queue:
                    queue_list.append({
                        'position': item[0],
                        'track_id': str(item[1]),
                        'title': item[2],
                        'artist': item[3],
                        'image_url': item[4],
                        'duration': item[5],
                        'added_by': item[6],
                        'added_at': item[7]
                    })
                
                return queue_list
                
        except Exception as e:
            logging.error(f"Error fetching room queue: {e}")
            return []
    
    def add_to_queue(self, room_id: str, track_id: uuid.UUID, user_id: str) -> bool:
        try:
            with SynqItDB() as db:
                db.cursor.execute(
                    """
                    SELECT COALESCE(MAX(position), 0) + 1
                    FROM room_queue
                    WHERE room_id = %s
                    """,
                    (room_id,)
                )
                next_position = db.cursor.fetchone()[0]
                
                db.cursor.execute(
                    """
                    INSERT INTO room_queue (id, room_id, position, track_id, added_by)
                    VALUES (%s, %s, %s, %s, %s)
                    """,
                    (str(uuid.uuid4()), room_id, next_position, track_id, user_id)
                )
                
                db.connection.commit()
                
                self.redis_manager.publish_room_update(room_id, {
                    'type': 'queue_update',
                    'room_id': room_id,
                    'action': 'add',
                    'position': next_position,
                    'track_id': str(track_id),
                    'added_by': user_id,
                    'timestamp': time.time()
                })
                
                return True
                
        except Exception as e:
            logging.error(f"Failed to add track to queue: {e}")
            return False
    
    def remove_from_queue(self, room_id: str, position: int) -> bool:
        try:
            with SynqItDB() as db:
                db.cursor.execute(
                    """
                    DELETE FROM room_queue
                    WHERE room_id = %s AND position = %s
                    """,
                    (room_id, position)
                )
                
                db.cursor.execute(
                    """
                    UPDATE room_queue
                    SET position = position - 1
                    WHERE room_id = %s AND position > %s
                    """,
                    (room_id, position)
                )
                
                db.connection.commit()
                
                self.redis_manager.publish_room_update(room_id, {
                    'type': 'queue_update',
                    'room_id': room_id,
                    'action': 'remove',
                    'position': position,
                    'timestamp': time.time()
                })
                
                return True
                
        except Exception as e:
            logging.error(f"Failed to remove track from queue: {e}")
            return False
            
    def generate_room_code(self, length=6):
        timestamp = str(time.time_ns())
        hash_digest = hashlib.sha256(timestamp.encode()).hexdigest().upper()
        clean_chars = ''.join(c for c in hash_digest if c not in '0O1I')
        return clean_chars[:length]
        
    def get_room_id_by_code(self, room_code: str) -> str:
        try:
            with SynqItDB() as db:
                db.cursor.execute(
                    f"""
                    SELECT id FROM {self.table_name}
                    WHERE room_code = %s AND status = 'ACTIVE'
                    """,
                    (room_code,)
                )
                result = db.cursor.fetchone()
                return result[0] if result else None
        except Exception as e:
            logging.error(f"Error fetching room ID by code: {e}")
            return None

    def get_room_code(self, room_id: str) -> str:
        try:
            with SynqItDB() as db:
                db.cursor.execute(
                    f"""
                    SELECT room_code FROM {self.table_name}
                    WHERE id = %s
                    """,
                    (room_id,)
                )
                result = db.cursor.fetchone()
                return result[0] if result else None
        except Exception as e:
            logging.error(f"Error fetching room code: {e}")
            return None