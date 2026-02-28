package party

import (
	"context"
	"encoding/json"

	pb "github.com/dis70rt/bluppi-backend/internals/gen/party"

	roompb "github.com/dis70rt/bluppi-backend/internals/gen/rooms"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GrpcHandler struct {
	service *Service
	pb.UnimplementedSyncServiceServer
	roompb.UnimplementedRoomServiceServer
}

func NewGrpcHandler(s *Service) *GrpcHandler {
	return &GrpcHandler{service: s}
}

func (h *GrpcHandler) ClockSync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	go func(roomID, userID string) {
        _ = h.service.RecordHeartbeat(context.Background(), roomID, userID)
    }(req.RoomId, req.UserId)
    
    return &pb.SyncResponse{
		ServerReceiveUs: CaptureServerReceiveUs(),
		ServerSendUs:    CaptureServerSendUs(),
	}, nil
}

func (h *GrpcHandler) CreateRoom(ctx context.Context, req *roompb.CreateRoomRequest) (*roompb.Room, error) {
    room, err := h.service.CreateRoom(
        ctx,
        req.Name,
        RoomVisibilityFromProto(req.Visibility),
        req.InviteOnly,
        req.HostUserId,
    )
    if err != nil {
        return nil, h.mapError(err)
    }
    return mapRoomToProto(room), nil
}

func (h *GrpcHandler) GetRoom(ctx context.Context, req *roompb.GetRoomRequest) (*roompb.Room, error) {
    room, err := h.service.GetRoom(ctx, req.RoomId)
    if err != nil {
        return nil, h.mapError(err)
    }
    return mapRoomToProto(room), nil
}

func (h *GrpcHandler) JoinRoom(ctx context.Context, req *roompb.JoinRoomRequest) (*emptypb.Empty, error) {
    var roomID string
    switch id := req.Identifier.(type) {
    case *roompb.JoinRoomRequest_RoomId:
        roomID = id.RoomId
    case *roompb.JoinRoomRequest_RoomCode:
        return nil, status.Error(codes.Unimplemented, "join by room code not implemented")
    default:
        return nil, status.Error(codes.InvalidArgument, "no room identifier provided")
    }
    if err := h.service.JoinRoom(ctx, roomID, req.UserId); err != nil {
        return nil, h.mapError(err)
    }
    return &emptypb.Empty{}, nil
}

func (h *GrpcHandler) LeaveRoom(ctx context.Context, req *roompb.LeaveRoomRequest) (*emptypb.Empty, error) {
    if err := h.service.LeaveRoom(ctx, req.RoomId, req.UserId); err != nil {
        return nil, h.mapError(err)
    }
    return &emptypb.Empty{}, nil
}

func (h *GrpcHandler) SubscribeToRoomEvents(req *roompb.SubscribeRequest, stream roompb.RoomService_SubscribeToRoomEventsServer) error {
    roomID := req.RoomId
    ctx := stream.Context()

    pubsub := h.service.redisRepo.SubscribeToRoom(ctx, roomID)
    defer pubsub.Close()
    ch := pubsub.Channel()

    for {
        select {
        case <-ctx.Done():
            _ = h.service.LeaveRoom(context.Background(), roomID, req.UserId)
            return nil

        case msg := <-ch:
            var rawEvent map[string]interface{}
            if err := json.Unmarshal([]byte(msg.Payload), &rawEvent); err != nil {
                continue
            }

            eventType := roompb.RoomEventType_ROOM_EVENT_TYPE_UNSPECIFIED
            if rawEvent["type"] == "USER_JOINED" {
                eventType = roompb.RoomEventType_ROOM_EVENT_TYPE_USER_JOINED
            } else if rawEvent["type"] == "USER_LEFT" {
                eventType = roompb.RoomEventType_ROOM_EVENT_TYPE_USER_LEFT
            }

            countFloat, _ := rawEvent["listener_count"].(float64)
            targetUserID, _ := rawEvent["target_user_id"].(string)

            roomEvent := &roompb.RoomEvent{
                Type:          eventType,
                RoomId:        roomID,
                ListenerCount: int32(countFloat),
                TargetUserId:  targetUserID,
            }

            if err := stream.Send(roomEvent); err != nil {
                return err
            }
        }
    }
}

// func (h *GrpcHandler) ListRooms(ctx context.Context, req *roompb.ListRoomsRequest) (*roompb.ListRoomsResponse, error) {
//     // Simple offset-based pagination for now
//     limit := int(req.PageSize)
//     if limit <= 0 {
//         limit = 20
//     }
//     offset := 0
//     // You can implement page_token parsing for real cursor-based paging if needed

//     summaries, err := h.service.repo.ListRooms(ctx, RoomVisibilityFromProto(req.Visibility), limit, offset)
//     if err != nil {
//         return nil, h.mapError(err)
//     }
//     return &roompb.ListRoomsResponse{
//         Rooms: mapRoomSummariesToProto(summaries),
//         NextPageToken: "",
//     }, nil
// }

// func (h *GrpcHandler) SearchRooms(ctx context.Context, req *roompb.SearchRoomsRequest) (*roompb.SearchRoomsResponse, error) {
//     return &roompb.SearchRoomsResponse{
//         Rooms: []*roompb.RoomSummary{},
//         NextPageToken: "",
//     }, nil
// }
// --- Error Mapping ---

func (h *GrpcHandler) mapError(err error) error {
    switch err {
    case ErrRoomNotFound:
        return status.Error(codes.NotFound, err.Error())
    case ErrInvalidInput:
        return status.Error(codes.InvalidArgument, err.Error())
    case ErrRoomCodeConflict:
        return status.Error(codes.AlreadyExists, err.Error())
    case ErrAlreadyInRoom:
        return status.Error(codes.AlreadyExists, err.Error())
    case ErrNotInRoom:
        return status.Error(codes.NotFound, err.Error())
    default:
        return status.Error(codes.Internal, err.Error())
    }
}

// --- Proto Enum Conversion ---

func RoomVisibilityFromProto(v roompb.RoomVisibility) RoomVisibility {
    switch v {
    case roompb.RoomVisibility_ROOM_VISIBILITY_PUBLIC:
        return RoomVisibilityPublic
    case roompb.RoomVisibility_ROOM_VISIBILITY_PRIVATE:
        return RoomVisibilityPrivate
    default:
        return ""
    }
}