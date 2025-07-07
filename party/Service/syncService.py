import grpc
import logging
import time
from rich import print_json

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

    def SendHostCommand(self, request_iterator, context):
        for request in request_iterator:
            print(f"Request: {request}")
            yield streaming_pb2.ServerResponse(
                error_message="ok",
                member_statuses=[],
                ready_member_count=100,
                total_member_count=50
            )
    
    def TimingSync(self, request, context):
        server_receive_time = int(time.time() * 1000)
        server_send_time = int(time.time() * 1000)

        return streaming_pb2.SyncReply(
            server_receive_ms=server_receive_time,
            server_send_ms=server_send_time
        )

    def MemberSync(self, request, context):
        pass
    



    