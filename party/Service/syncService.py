import grpc
import logging
import time
from rich import print_json
from collections import defaultdict
import threading

from protobuf import streaming_pb2_grpc, streaming_pb2

from Manager.roomManager import RoomManager
from Manager.redisManager import RedisManager

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)

class SyncService(streaming_pb2_grpc.SyncService):
    def __init__(self):
        self.redis_manager = RedisManager()
        self.room_manager = RoomManager()
        self.member_streams = defaultdict(dict)
        self.stream_lock = threading.Lock()

    def SendHostCommand(self, request_iterator, context):
        try:
            for request in request_iterator:
                room_id = None
                if request.HasField('track_command'):
                    room_id = request.track_command.room_id
                elif request.HasField('position_update'):
                    room_id = request.position_update.room_id
                elif request.HasField('control_command'):
                    room_id = request.control_command.room_id

                if not room_id:
                    logging.WARNING("No Room ID found in host_command")
                    yield streaming_pb2.ServerResponse(
                        type=streaming_pb2.ServerResponse.ERROR,
                        error_message="No Room ID found in host_command",
                    )
                    continue

                room_members = self.redis_manager.get_room_members(room_id=room_id)
                if not room_members:
                    logging.WARNING("No members found for room {room_id}")
                    yield streaming_pb2.ServerResponse(
                        type=streaming_pb2.ServerResponse.ERROR,
                        error_message="No members found for room {room_id}",
                    )
                    continue

                broadcast = streaming_pb2.ServerBroadcast(
                    room_id=room_id
                )

                if request.HasField('position_update'):
                    broadcast.type = streaming_pb2.ServerBroadcast.POSITION_UPDATE
                    broadcast.position_update.CopyFrom(request.position_update)
                elif request.HasField('control_command'):
                    broadcast.type = streaming_pb2.ServerBroadcast.CONTROL_COMMAND
                    broadcast.control_command.CopyFrom(request.control_command)
                elif request.HasField('track_command'):
                    broadcast.type = streaming_pb2.ServerBroadcast.TRACK_COMMAND
                    broadcast.track_command.CopyFrom(request.track_command)

                self.broadcast_to_room_members(room_id, broadcast)

                yield streaming_pb2.ServerResponse(
                    type=streaming_pb2.ServerResponse.ACKNOWLEDGED,
                    ready_member_count=len(room_members),
                    total_member_count=len(room_members),
                    error_message="Command broadcasted successfully"
                )

        except Exception as e:
            logging.error(f"Error in SendHostCommand: {e}")
            yield streaming_pb2.ServerResponse(
                type=streaming_pb2.ServerResponse.ERROR,
                error_message=f"Internal server error: {str(e)}"
            )

    def broadcast_to_room_members(self, room_id, broadcast):
         with self.stream_lock:
            if room_id not in self.member_streams:
                logging.warning(f"No active streams for room {room_id}")
                return
            
            dead_streams = []
            for user_id, stream in self.member_streams[room_id].items():
                try:
                    logging.info(f"Broadcasting to user {user_id[:8]}... in room {room_id[:8]}...")
                    stream.put(broadcast)
                except Exception as e:
                    logging.error(f"Failed to send broadcast to user {user_id}: {e}")
                    dead_streams.append(user_id)
            
            for user_id in dead_streams:
                del self.member_streams[room_id][user_id]
                
            if not self.member_streams[room_id]:
                del self.member_streams[room_id]

    
    def TimingSync(self, request, context):
        server_receive_time = int(time.time() * 1000)
        server_send_time = int(time.time() * 1000)

        return streaming_pb2.SyncReply(
            server_receive_ms=server_receive_time,
            server_send_ms=server_send_time
        )

    def MemberSync(self, request_iterator, context):
        import queue
        import threading

        broadcast_queue = queue.Queue()
        user_id = None
        room_id = None

        def read_client():
            nonlocal user_id, room_id
            try:
                for req in request_iterator:
                    if user_id is None:
                        user_id = req.user_id
                        room_id = req.room_id
                        with self.stream_lock:
                            self.member_streams.setdefault(room_id, {})[user_id] = broadcast_queue

                        logging.info(f"Member {user_id[:8]} connected to room {room_id[:8]}")

                    logging.info(f"Received member status from {user_id[:8]}: {req.status}")
                    # TODO: store position_ms, etc., into Redis

            except Exception as e:
                logging.error(f"Error reading client stream: {e}")
            finally:
                broadcast_queue.put(None)

        reader = threading.Thread(target=read_client, daemon=True)
        reader.start()

        try:
            while True:
                item = broadcast_queue.get()
                if item is None:
                    break
                yield item

        except Exception as e:
            logging.error(f"Error in broadcast sender: {e}")

        finally:
            if user_id and room_id:
                with self.stream_lock:
                    if (
                        room_id in self.member_streams
                        and user_id in self.member_streams[room_id]
                    ):
                        del self.member_streams[room_id][user_id]
                        if not self.member_streams[room_id]:
                            del self.member_streams[room_id]

                logging.info(f"Member {user_id[:8]} disconnected from room {room_id[:8]}")
