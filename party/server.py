import grpc
from concurrent import futures

from protobuf import room_pb2
from protobuf import room_pb2_grpc
from roomManager import RoomManager

class ListeningParty(room_pb2_grpc.RoomServiceServicer):
    def __init__(self):
        self.room_manager = RoomManager()
    
    def CreateRoom(self, request, context):
        try:
            room_id = self.room_manager.create_room(
                name=request.name,
                host_user_id=request.host_user_id,
                description=request.description,
                visibility=request.visibility.name,  # Convert enum to string
                invite_only=request.invite_only
            )
            
            return room_pb2.Room(
                id=room_id,
                name=request.name,
                description=request.description,
                host_user_id=request.host_user_id,
                visibility=request.visibility,
                invite_only=request.invite_only
            )

        except Exception as e:
            context.set_details(f"Failed to create room: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            return room_pb2.Room()

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    room_pb2_grpc.add_RoomServiceServicer_to_server(ListeningParty(), server)
    listen_addr = '[::]:50051'
    server.add_insecure_port(listen_addr)
    print(f"Starting server on {listen_addr}")
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    serve()