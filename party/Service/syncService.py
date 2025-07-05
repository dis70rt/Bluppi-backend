import grpc
import logging
import time
from google.protobuf import empty_pb2 as google_pb2

from protobuf import playback_pb2_grpc
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

    def MeasureTiming(self, request, context):
        pass
    
    def Sync(self, request, context):
        server_receive_time = int(time.time() * 1000)
        server_send_time = int(time.time() * 1000)

        return streaming_pb2.SyncReply(
            server_receive_ms=server_receive_time,
            server_send_ms=server_send_time
        )

    def SyncPlayback(self, request, context):
        pass
    



    