import redis
import json
import logging
import time
import uuid
from typing import Set

import os
from dotenv import load_dotenv
load_dotenv(override=True)

redis_host = os.getenv("REDIS_HOST", "localhost")
redis_port = int(os.getenv("REDIS_PORT", 6379))
redis_db   = int(os.getenv("REDIS_DB",   0))

class RedisManager:
    def __init__(self):
        self.redis_client = redis.Redis(
            host=redis_host,
            port=redis_port,
            decode_responses=True
        )
        self.pubsub = self.redis_client.pubsub()
        
    def create_room_channel(self, room_id: str) -> bool:
        try:
            self.redis_client.hset(f'room:{room_id}:info', 'status', 'ACTIVE')
            self.redis_client.hset(f'room:{room_id}:info', 'created_at', str(time.time()))
            self.redis_client.hset(f'room:{room_id}:info', 'last_activity', str(time.time()))
            
            self.redis_client.delete(f'room:{room_id}:members')
            
            self.redis_client.hset(f'room:{room_id}:playback', 'status', 'PAUSED')
            self.redis_client.hset(f'room:{room_id}:playback', 'track_id', '')
            self.redis_client.hset(f'room:{room_id}:playback', 'position_ms', '0')
            self.redis_client.hset(f'room:{room_id}:playback', 'updated_at', str(time.time()))
            
            logging.info(f"Created room channel for {room_id}")
            return True
        except Exception as e:
            logging.error(f"Failed to create room channel {room_id}: {e}")
            return False
    
    def is_room_active(self, room_id: str) -> bool:
        status = self.redis_client.hget(f'room:{room_id}:info', 'status')
        return status == 'ACTIVE'
    
    def update_room_activity(self, room_id: str):
        self.redis_client.hset(f'room:{room_id}:info', 
                              'last_activity', str(time.time()))
    
    def set_host_connected(self, room_id: str, user_id: str):
        self.redis_client.hset(f'room:{room_id}:host', 'user_id', user_id)
        self.redis_client.hset(f'room:{room_id}:host', 'connected', 'True')
        self.redis_client.hset(f'room:{room_id}:host', 'last_seen', str(time.time()))
        self.update_room_activity(room_id)
    
    def set_host_disconnected(self, room_id: str):
        self.redis_client.hset(f'room:{room_id}:host', 'connected', 'False')
        self.redis_client.hset(f'room:{room_id}:host', 'disconnected_at', str(time.time()))
        self.redis_client.hset(f'room:{room_id}:info', 'status', 'INACTIVE')
    
    def is_host_connected(self, room_id: str) -> bool:
        connected = self.redis_client.hget(f'room:{room_id}:host', 'connected')
        return connected == 'True'
    
    def add_member_to_room(self, room_id: str, user_id: str) -> int:
        self.redis_client.sadd(f'room:{room_id}:members', user_id)
        self.redis_client.sadd(f'user:{user_id}:rooms', room_id)
        return self.redis_client.scard(f'room:{room_id}:members')
    
    def remove_member_from_room(self, room_id: str, user_id: str) -> int:
        self.redis_client.srem(f'room:{room_id}:members', user_id)
        self.redis_client.srem(f'user:{user_id}:rooms', room_id)
        return self.redis_client.scard(f'room:{room_id}:members')
    
    def get_room_members(self, room_id: str) -> Set[str]:
        return self.redis_client.smembers(f'room:{room_id}:members')
    
    def get_member_count(self, room_id: str) -> int:
        return self.redis_client.scard(f'room:{room_id}:members')
    
    def publish_room_update(self, room_id: str, update_data: dict):
        channel = f'room:{room_id}:updates'
        if 'track_id' in update_data and isinstance(update_data['track_id'], uuid.UUID):
            update_data['track_id'] = str(update_data['track_id'])
        message = json.dumps(update_data)
        result = self.redis_client.publish(channel, message)
        self.update_room_activity(room_id)
        logging.info(f"Published to {channel}: {message} (subscribers: {result})")
    
    def subscribe_to_room(self, room_id: str):
        channel = f'room:{room_id}:updates'
        self.pubsub.subscribe(channel)
        logging.info(f"Subscribed to {channel}")
    
    def get_room_snapshot(self, room_id: str) -> dict:
        return {
            'room_info': self.redis_client.hgetall(f'room:{room_id}:info'),
            'playback_state': self.redis_client.hgetall(f'room:{room_id}:playback'),
            'members': list(self.redis_client.smembers(f'room:{room_id}:members')),
            'member_count': self.redis_client.scard(f'room:{room_id}:members'),
            'host_info': self.redis_client.hgetall(f'room:{room_id}:host')
        }
        
    def update_playback_state(self, room_id: str, updates: dict):
        for key, value in updates.items():
            if key == 'track_id' and isinstance(value, uuid.UUID):
                value = str(value)
            self.redis_client.hset(f'room:{room_id}:playback', key, value)
        self.redis_client.hset(f'room:{room_id}:playback', 'updated_at', str(time.time()))
        
        self.publish_room_update(room_id, {
            'type': 'playback_update',
            'room_id': room_id,
            'changes': updates,
            'timestamp': time.time()
        })

    def client_offsets(self, client_id, offset):
        key = f"offsets:{client_id}"
        return self.redis_client.rpush(key,offset)