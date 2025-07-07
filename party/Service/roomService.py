import grpc
import logging

from protobuf import room_pb2
from protobuf import room_pb2_grpc
from protobuf import common_pb2
from google.protobuf import empty_pb2 as google_pb2

from Manager.roomManager import RoomManager

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)

class RoomService(room_pb2_grpc.RoomServiceServicer):
    def __init__(self):
        self.room_manager = RoomManager()
    
    def CreateRoom(self, request, context):
        try:
            if request.visibility == 0:
                visibility = "PUBLIC"
            else: 
                visibility = "PRIVATE"
                
            room_id = self.room_manager.create_room(
                name=request.name,
                host_user_id=request.host_user_id,
                description=request.description,
                visibility=visibility, 
            )
            
            room_code = self.room_manager.get_room_code(room_id)
            
            return room_pb2.Room(
                id=room_id,
                name=request.name,
                description=request.description,
                room_code=room_code,
                host_user_id=request.host_user_id,
                visibility=request.visibility,
                invite_only=False,
                status=common_pb2.ACTIVE
            )

        except Exception as e:
            logging.error(f"Failed to create room: {e}")
            context.set_details(f"Failed to create room: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            return room_pb2.Room()
        
    def ListRooms(self, request, context):
        try:
            page_size = request.page_size if request.page_size > 0 else 50
            
            visibility_filter = None
            if request.HasField('visibility_filter'):
                visibility_filter = "PUBLIC" if request.visibility_filter == common_pb2.PUBLIC else "PRIVATE"
            
            host_user_id_filter = None
            if request.HasField('host_user_id_filter'):
                host_user_id_filter = request.host_user_id_filter
            
            include_private = False
            if request.HasField('include_private_rooms_if_member'):
                include_private = request.include_private_rooms_if_member
            
            logging.info(f"Listing active rooms with page_size={page_size}, visibility={visibility_filter}")
            
            rooms = self.room_manager.list_active_rooms(
                visibility_filter=visibility_filter,
                host_user_id_filter=host_user_id_filter,
                include_private_rooms_if_member=include_private,
            )
            
            response = room_pb2.ListRoomsResponse()
            for room_data in rooms:
                room = response.rooms.add()
                room.id = room_data['id']
                room.name = room_data['name']
                room.description = room_data['description']
                room.room_code = room_data['room_code']
                room.host_user_id = room_data['host_user_id']
                room.visibility = common_pb2.PUBLIC if room_data['visibility'] == 'PUBLIC' else common_pb2.PRIVATE
                room.invite_only = False
                room.status = common_pb2.ACTIVE
            
            return response
            
        except Exception as e:
            logging.error(f"Failed to list rooms: {e}")
            context.set_details(f"Failed to list rooms: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            return room_pb2.ListRoomsResponse()

    def JoinRoom(self, request, context):
        try:
            if request.HasField('room_id'):
                room_id = request.room_id
            else:
                room_code = request.room_code
                room_id = self.room_manager.get_room_id_by_code(room_code)
                if not room_id:
                    context.set_code(grpc.StatusCode.NOT_FOUND)
                    context.set_details(f"Room with code {room_code} not found")
                    return room_pb2.Room()
                    
            user_id = request.user_id
            
            logging.info(f"User {user_id[:8]}... joining room {room_id[:8]}...")
            
            room_info = self.room_manager.get_room_info(room_id)
            if not room_info:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details(f"Room {room_id} not found")
                return room_pb2.Room()
                
            if room_info['status'] != 'ACTIVE':
                context.set_code(grpc.StatusCode.FAILED_PRECONDITION)
                context.set_details(f"Room {room_id} is not active")
                return room_pb2.Room()
                
            success = self.room_manager.join_room(room_id, user_id)
            
            if not success:
                context.set_code(grpc.StatusCode.INTERNAL)
                context.set_details("Failed to join room")
                return room_pb2.Room()
                
            return room_pb2.Room(
                id=room_info['id'],
                name=room_info['name'],
                description=room_info['description'],
                room_code=room_info['room_code'],
                host_user_id=room_info['host_user_id'],
                visibility=common_pb2.PUBLIC if room_info['visibility'] == 'PUBLIC' else common_pb2.PRIVATE,
                invite_only=False,
                status=common_pb2.ACTIVE
            )
            
        except Exception as e:
            logging.error(f"Failed to join room: {e}")
            context.set_details(f"Failed to join room: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            return room_pb2.Room()
        
    def LeaveRoom(self, request, context):
        try:
            room_id = request.room_id
            user_id = request.user_id
            
            logging.info(f"User {user_id[:8]}... leaving room {room_id[:8]}...")
            
            room_info = self.room_manager.get_room_info(room_id)
            if not room_info:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details(f"Room {room_id} not found")
                return google_pb2.Empty()
                
            success = self.room_manager.leave_room(room_id, user_id)
            
            if not success:
                context.set_code(grpc.StatusCode.INTERNAL)
                context.set_details("Failed to leave room")
                return google_pb2.Empty()
                
            return google_pb2.Empty()
            
        except Exception as e:
            logging.error(f"Failed to leave room: {e}")
            context.set_details(f"Failed to leave room: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            return google_pb2.Empty()


