import grpc
import time
import logging
from google.protobuf import timestamp_pb2

from protobuf import room_pb2
from protobuf import streaming_pb2
from protobuf import streaming_pb2_grpc
from protobuf import common_pb2
from protobuf import playback_pb2

from Manager.roomManager import RoomManager
from Manager.redisManager import RedisManager

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)

class RoomStreamService(streaming_pb2_grpc.RoomStreamServiceServicer):
    def __init__(self):
        self.redis_manager = RedisManager()
        self.room_manager = RoomManager()
        self.active_streams = {}
        
    def JoinRoomStream(self, request, context):
        try:
            room_id = request.room_id
            user_id = request.user_id
            
            logging.info(f"User {user_id[:8]}... joining stream for room {room_id[:8]}...")
            
            if not self.redis_manager.is_room_active(room_id):
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details("Room not found or inactive")
                return streaming_pb2.RoomStreamSnapshot()
            
            room_info = self.room_manager.get_room_info(room_id)
            if not room_info:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details(f"Room {room_id} not found in database")
                return streaming_pb2.RoomStreamSnapshot()
            
            if not self.room_manager.join_room(room_id, user_id):
                context.set_code(grpc.StatusCode.INTERNAL)
                context.set_details("Failed to join room")
                return streaming_pb2.RoomStreamSnapshot()
            
            snapshot_data = self.redis_manager.get_room_snapshot(room_id)
            members = self.redis_manager.get_room_members(room_id)
            member_objects = []
            for member_id in members:
                member_objects.append(room_pb2.RoomMember(
                    user_id=member_id,
                    role=common_pb2.PARTICIPANT if member_id != room_info['host_user_id'] else common_pb2.HOST
                ))

            print("Members in room:", members)
            snapshot = streaming_pb2.RoomStreamSnapshot(
                room_info=room_pb2.Room(
                    id=room_info['id'],
                    name=room_info['name'],
                    description=room_info['description'],
                    room_code=room_info['room_code'],
                    host_user_id=room_info['host_user_id'],
                    visibility=common_pb2.PUBLIC if room_info['visibility'] == 'PUBLIC' else common_pb2.PRIVATE,
                    invite_only=room_info['invite_only'],
                    status=common_pb2.ACTIVE,
                    members=member_objects
                ),
                current_playback=playback_pb2.PlaybackState(
                    room_id=room_id,
                    current_track_id=snapshot_data['playback_state'].get('track_id', ''),
                    position_ms=int(snapshot_data['playback_state'].get('position_ms', 0)),
                    status=common_pb2.PLAYING if snapshot_data['playback_state'].get('status') == 'PLAYING' else common_pb2.PAUSED
                ),
                member_count=snapshot_data['member_count']
            )
            
            logging.info(f"Sent snapshot to user {user_id[:8]}...: {snapshot_data['member_count']} members")
            return snapshot
            
        except Exception as e:
            logging.error(f"Error joining room stream: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return streaming_pb2.RoomStreamSnapshot()
    
    def StreamRoomUpdates(self, request, context):
        room_id = request.room_id
        user_id = request.user_id
        
        logging.info(f"Starting stream for user {user_id[:8]}... in room {room_id[:8]}...")        
        pubsub = None
        
        try:
            if room_id not in self.active_streams:
                self.active_streams[room_id] = set()
            self.active_streams[room_id].add(context)
            
            pubsub = self.redis_manager.redis_client.pubsub()
            channel = f'room:{room_id}:updates'
            pubsub.subscribe(channel)
            
            logging.info(f"Subscribed to channel: {channel}")
            
            while context.is_active():
                try:
                    message = pubsub.get_message(timeout=1.0)
                    
                    if message and message['type'] == 'message':
                        update = self.parse_redis_message_to_proto(message)
                        if update:
                            logging.info(f"Sending update to {user_id[:8]}...: {message['data'][:100]}...")
                            yield update
                            
                except grpc.RpcError as e:
                    logging.error(f"gRPC error in stream for {user_id[:8]}...: {e}")
                    break
                except Exception as e:
                    logging.error(f"Error in stream loop for {user_id[:8]}...: {e}")
                    break
            
            logging.info(f"Stream ended for user {user_id[:8]}...")
            
        except Exception as e:
            logging.error(f"Stream setup error for {user_id[:8]}...: {e}")
        finally:
            logging.info(f"Cleaning up stream for user {user_id[:8]}...")
            
            try:
                if room_id in self.active_streams:
                    self.active_streams[room_id].discard(context)
                    if not self.active_streams[room_id]:
                        del self.active_streams[room_id]
            except:
                pass
            
            try:
                self.room_manager.leave_room(room_id, user_id)
            except Exception as e:
                logging.error(f"Error removing user from room: {e}")
            
            try:
                if pubsub:
                    pubsub.close()
            except Exception as e:
                logging.error(f"Error closing pubsub: {e}")
    
    def parse_redis_message_to_proto(self, message):
        try:
            import json
            data = json.loads(message['data'])
            update_type = data.get('type')
            
            timestamp = timestamp_pb2.Timestamp()
            timestamp.FromSeconds(int(data.get('timestamp', time.time())))
            
            room_update = streaming_pb2.RoomStreamUpdate(
                room_id=data['room_id'],
                timestamp=timestamp
            )
            
            if update_type == 'member_join':
                room_update.member_update.member_join.user.id = data['user_id']
                room_update.member_update.member_join.role = common_pb2.PARTICIPANT
                
            elif update_type == 'member_leave':
                room_update.member_update.member_leave.user_id = data['user_id']
                
            elif update_type == 'playback_update':
                changes = data.get('changes', {})
                if 'track_id' in changes:
                    room_update.playback_update.track_change.current_track.track_id = changes['track_id']
                    room_update.playback_update.track_change.position_ms = int(changes.get('position_ms', 0))
                elif 'status' in changes:
                    status = common_pb2.PLAYING if changes['status'] == 'PLAYING' else common_pb2.PAUSED
                    room_update.playback_update.play_state.status = status
                    room_update.playback_update.play_state.position_ms = int(changes.get('position_ms', 0))
                elif 'position_ms' in changes:
                    room_update.playback_update.seek.position_ms = int(changes['position_ms'])
            
            elif update_type == 'host_disconnected':
                room_update.room_status_update.status = common_pb2.INACTIVE
                room_update.room_status_update.reason = "host_disconnected"
            
            return room_update
            
        except Exception as e:
            logging.error(f"Error parsing Redis message: {e}")
            return None