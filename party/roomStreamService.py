import grpc
import asyncio
import json
import logging
from concurrent import futures
from typing import Dict, Set

from protobuf import stream_pb2
from protobuf import stream_pb2_grpc
from protobuf import room_pb2
from protobuf import track_pb2
from protobuf import user_pb2
from protobuf import common_pb2

from redisManager import RedisManager
from roomManager import RoomManager

class RoomStreamService(stream_pb2_grpc.RoomStreamServiceServicer):
    def __init__(self):
        self.redis_manager = RedisManager()
        self.room_manager = RoomManager()
        self.active_streams: Dict[str, Set[grpc.ServicerContext]] = {}
        
        # Start background task for Redis subscription
        asyncio.create_task(self.listen_to_redis_updates())
    
    def JoinRoomStream(self, request, context):
        """Get initial room snapshot when joining stream"""
        try:
            room_id = request.room_id
            user_id = request.user_id
            
            # Check if room exists and is active
            if not self.redis_manager.is_room_active(room_id):
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details("Room not found or inactive")
                return stream_pb2.RoomStreamSnapshot()
            
            # Add user to room
            self.room_manager.join_room(room_id, user_id)
            
            # Get current room snapshot
            snapshot_data = self.redis_manager.get_room_snapshot(room_id)
            
            # TODO: Fetch actual user/track data from database
            # For now, returning basic snapshot
            snapshot = stream_pb2.RoomStreamSnapshot(
                room_info=room_pb2.Room(
                    id=room_id,
                    # ... populate from database
                ),
                current_playback=playback_pb2.PlaybackState(
                    room_id=room_id,
                    current_track_id=snapshot_data['playback_state'].get('track_id', ''),
                    position_ms=int(snapshot_data['playback_state'].get('position_ms', 0)),
                    status=common_pb2.PLAYING if snapshot_data['playback_state'].get('status') == 'PLAYING' else common_pb2.PAUSED
                ),
                member_count=snapshot_data['member_count']
            )
            
            return snapshot
            
        except Exception as e:
            logging.error(f"Error joining room stream: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return stream_pb2.RoomStreamSnapshot()
    
    def StreamRoomUpdates(self, request, context):
        """Stream real-time updates for a room"""
        room_id = request.room_id
        user_id = request.user_id
        
        try:
            # Add this stream to active streams
            if room_id not in self.active_streams:
                self.active_streams[room_id] = set()
            self.active_streams[room_id].add(context)
            
            # Subscribe to Redis updates for this room
            self.redis_manager.subscribe_to_room(room_id)
            
            # Keep stream alive
            while not context.is_active():
                try:
                    # Process Redis messages and yield updates
                    message = self.redis_manager.pubsub.get_message(timeout=1.0)
                    if message and message['type'] == 'message':
                        update = self.parse_redis_message_to_proto(message)
                        if update:
                            yield update
                            
                except Exception as e:
                    logging.error(f"Error in stream: {e}")
                    break
            
        except Exception as e:
            logging.error(f"Stream error: {e}")
        finally:
            # Clean up
            if room_id in self.active_streams:
                self.active_streams[room_id].discard(context)
                if not self.active_streams[room_id]:
                    del self.active_streams[room_id]
            
            # Remove user from room
            self.room_manager.leave_room(room_id, user_id)
    
    def parse_redis_message_to_proto(self, message) -> stream_pb2.RoomStreamUpdate:
        """Convert Redis message to protobuf update"""
        try:
            data = json.loads(message['data'])
            update_type = data.get('type')
            
            room_update = stream_pb2.RoomStreamUpdate(
                room_id=data['room_id'],
                timestamp=self.create_timestamp(data['timestamp'])
            )
            
            if update_type == 'member_join':
                # Only send incremental change
                room_update.member_update.member_join.user.id = data['user_id']
                room_update.member_update.member_join.role = common_pb2.PARTICIPANT
                
            elif update_type == 'member_leave':
                room_update.member_update.member_leave.user_id = data['user_id']
                
            elif update_type == 'playback_update':
                changes = data['changes']
                if 'track_id' in changes:
                    room_update.playback_update.track_change.current_track.track_id = changes['track_id']
                if 'status' in changes:
                    status = common_pb2.PLAYING if changes['status'] == 'PLAYING' else common_pb2.PAUSED
                    room_update.playback_update.play_state.status = status
                if 'position_ms' in changes:
                    room_update.playback_update.play_state.position_ms = changes['position_ms']
            
            elif update_type == 'host_disconnected':
                room_update.room_status_update.status = common_pb2.INACTIVE
                room_update.room_status_update.reason = "host_disconnected"
            
            return room_update
            
        except Exception as e:
            logging.error(f"Error parsing Redis message: {e}")
            return None
    
    async def listen_to_redis_updates(self):
        """Background task to listen for Redis updates"""
        while True:
            try:
                # Process Redis pub/sub messages
                await asyncio.sleep(0.1)  # Small delay to prevent busy waiting
            except Exception as e:
                logging.error(f"Redis listener error: {e}")
                await asyncio.sleep(1)