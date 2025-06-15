import redis
import json
import asyncio
import logging
from typing import Dict, Set, Optional
import time

class RedisManager:
    def __init__(self):
        self.redis_client = redis.Redis(
            host='localhost',
            port=6379,
            decode_responses=True
        )
        self.pubsub = self.redis_client.pubsub()
        
    # Room channel management
    def create_room_channel(self, room_id: str) -> bool:
        """Create a new room channel"""
        try:
            # Set room as active
            self.redis_client.hset(f'room:{room_id}:info', {
                'status': 'ACTIVE',
                'created_at': time.time(),
                'last_activity': time.time()
            })
            
            # Initialize empty member set
            self.redis_client.delete(f'room:{room_id}:members')
            
            # Initialize playback state
            self.redis_client.hset(f'room:{room_id}:playback', {
                'status': 'PAUSED',
                'track_id': '',
                'position_ms': 0,
                'updated_at': time.time()
            })
            
            return True
        except Exception as e:
            logging.error(f"Failed to create room channel {room_id}: {e}")
            return False
    
    def is_room_active(self, room_id: str) -> bool:
        """Check if room channel is active"""
        status = self.redis_client.hget(f'room:{room_id}:info', 'status')
        return status == 'ACTIVE'
    
    def update_room_activity(self, room_id: str):
        """Update last activity timestamp"""
        self.redis_client.hset(f'room:{room_id}:info', 
                              'last_activity', time.time())
    
    # Host management
    def set_host_connected(self, room_id: str, user_id: str):
        """Mark host as connected"""
        self.redis_client.hset(f'room:{room_id}:host', {
            'user_id': user_id,
            'connected': True,
            'last_seen': time.time()
        })
        self.update_room_activity(room_id)
    
    def set_host_disconnected(self, room_id: str):
        """Mark host as disconnected and start countdown"""
        self.redis_client.hset(f'room:{room_id}:host', {
            'connected': False,
            'disconnected_at': time.time()
        })
        
        # Set expiration for 3 minutes
        self.redis_client.expire(f'room:{room_id}:host:timeout', 180)
        self.redis_client.set(f'room:{room_id}:host:timeout', 'waiting')
    
    def is_host_connected(self, room_id: str) -> bool:
        """Check if host is connected"""
        connected = self.redis_client.hget(f'room:{room_id}:host', 'connected')
        return connected == 'True'
    
    # Member management
    def add_member_to_room(self, room_id: str, user_id: str) -> int:
        """Add member to room and return new count"""
        self.redis_client.sadd(f'room:{room_id}:members', user_id)
        self.redis_client.sadd(f'user:{user_id}:rooms', room_id)
        return self.redis_client.scard(f'room:{room_id}:members')
    
    def remove_member_from_room(self, room_id: str, user_id: str) -> int:
        """Remove member from room and return new count"""
        self.redis_client.srem(f'room:{room_id}:members', user_id)
        self.redis_client.srem(f'user:{user_id}:rooms', room_id)
        return self.redis_client.scard(f'room:{room_id}:members')
    
    def get_room_members(self, room_id: str) -> Set[str]:
        """Get all room members"""
        return self.redis_client.smembers(f'room:{room_id}:members')
    
    def get_member_count(self, room_id: str) -> int:
        """Get current member count"""
        return self.redis_client.scard(f'room:{room_id}:members')
    
    # Pub/Sub for real-time updates
    def publish_room_update(self, room_id: str, update_data: dict):
        """Publish update to room channel"""
        channel = f'room:{room_id}:updates'
        message = json.dumps(update_data)
        self.redis_client.publish(channel, message)
        self.update_room_activity(room_id)
    
    def subscribe_to_room(self, room_id: str):
        """Subscribe to room updates"""
        channel = f'room:{room_id}:updates'
        self.pubsub.subscribe(channel)
    
    def get_room_snapshot(self, room_id: str) -> dict:
        """Get current room state snapshot"""
        return {
            'room_info': self.redis_client.hgetall(f'room:{room_id}:info'),
            'playback_state': self.redis_client.hgetall(f'room:{room_id}:playback'),
            'members': list(self.redis_client.smembers(f'room:{room_id}:members')),
            'member_count': self.redis_client.scard(f'room:{room_id}:members'),
            'host_info': self.redis_client.hgetall(f'room:{room_id}:host')
        }