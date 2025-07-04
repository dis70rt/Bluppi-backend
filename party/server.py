import grpc
import signal
import sys
import logging
from concurrent import futures

from Service.roomService import RoomService
from Service.roomStreamService import RoomStreamService
from protobuf import room_pb2_grpc
from protobuf import streaming_pb2_grpc

class BluppiServer:
    def __init__(self, port=50051):
        self.port = port
        self.server = None
        
    def start(self):
        self.server = grpc.server(
            futures.ThreadPoolExecutor(max_workers=50),
            options=[
                ('grpc.keepalive_time_ms', 30000),
                ('grpc.keepalive_timeout_ms', 10000),
                ('grpc.keepalive_permit_without_calls', True),
                ('grpc.http2.max_pings_without_data', 0),
                ('grpc.http2.min_time_between_pings_ms', 10000),
                ('grpc.http2.min_ping_interval_without_data_ms', 300000),
                ('grpc.max_send_message_length', 4 * 1024 * 1024),
                ('grpc.max_receive_message_length', 4 * 1024 * 1024),
            ]
        )
        
        room_pb2_grpc.add_RoomServiceServicer_to_server(RoomService(), self.server)
        streaming_pb2_grpc.add_RoomStreamServiceServicer_to_server(RoomStreamService(), self.server)
        
        listen_addr = f'[::]:{self.port}'
        self.server.add_insecure_port(listen_addr)
        
        logging.info(f"Starting Bluppi server on {listen_addr}")
        logging.info("Services available:")
        logging.info("  - RoomService (room management)")
        logging.info("  - RoomStreamService (real-time streaming)")
        
        self.server.start()
        
        signal.signal(signal.SIGINT, self._signal_handler)
        signal.signal(signal.SIGTERM, self._signal_handler)
        
        try:
            self.server.wait_for_termination()
        except KeyboardInterrupt:
            logging.info("Received interrupt signal")
        finally:
            self.stop()
    
    def stop(self):
        if self.server:
            logging.info("Stopping server...")
            self.server.stop(grace=5)
            logging.info("Server stopped")
    
    def _signal_handler(self, signum, frame):
        logging.info(f"Received signal {signum}")
        self.stop()
        sys.exit(0)

def serve():
    server = BluppiServer()
    server.start()

if __name__ == '__main__':
    serve()