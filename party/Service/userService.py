from protobuf import user_pb2
from protobuf import user_pb2_grpc

class UserService(user_pb2_grpc.UserServiceServicer):
    def __init__(self, user_manager):
        self.user_manager = user_manager

    def GetUserProfile(self, request, context):
        try:
            user_profile = self.user_manager.get_user_profile(request.user_id)
            if not user_profile:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details(f"User {request.user_id} not found")
                return user_pb2.UserProfile()

            return user_pb2.UserProfile(
                user_id=user_profile.user_id,
                username=user_profile.username,
                email=user_profile.email,
                bio=user_profile.bio,
                profile_picture_url=user_profile.profile_picture_url
            )
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Error retrieving user profile: {str(e)}")
            return user_pb2.UserProfile()