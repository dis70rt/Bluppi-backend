package party

import (
	"context"
	"encoding/json"
	"strconv"

	pb "github.com/dis70rt/bluppi-backend/internals/gen/party"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/middlewares"

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
	userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, h.mapError(err)
    }

	go func(roomID, userID string) {
		_ = h.service.RecordHeartbeat(context.Background(), roomID, userID)
	}(req.RoomId, userID)

	return &pb.SyncResponse{
		ServerReceiveUs: CaptureServerReceiveUs(),
		ServerSendUs:    CaptureServerSendUs(),
	}, nil
}

func (h *GrpcHandler) CreateRoom(ctx context.Context, req *roompb.CreateRoomRequest) (*roompb.Room, error) {
	userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, h.mapError(err)
    }
	
	room, err := h.service.CreateRoom(
		ctx,
		req.Name,
		RoomVisibilityFromProto(req.Visibility),
		req.InviteOnly,
		userID,
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

	userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, h.mapError(err)
    }

	if err := h.service.JoinRoom(ctx, roomID, userID); err != nil {
		return nil, h.mapError(err)
	}
	return &emptypb.Empty{}, nil
}

func (h *GrpcHandler) LeaveRoom(ctx context.Context, req *roompb.LeaveRoomRequest) (*emptypb.Empty, error) {
	userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, h.mapError(err)
    }
	
	if err := h.service.LeaveRoom(ctx, req.RoomId, userID); err != nil {
		return nil, h.mapError(err)
	}
	return &emptypb.Empty{}, nil
}

func (h *GrpcHandler) SubscribeToRoomEvents(req *roompb.SubscribeRequest, stream roompb.RoomService_SubscribeToRoomEventsServer) error {
	roomID := req.RoomId
	ctx := stream.Context()

	userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return h.mapError(err)
    }

	pubsub := h.service.redisRepo.SubscribeToRoom(ctx, roomID)
	defer pubsub.Close()
	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			_ = h.service.LeaveRoom(context.Background(), roomID, userID)
			return nil

		case msg := <-ch:
			var rawEvent map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Payload), &rawEvent); err != nil {
				continue
			}

			roomEvent := &roompb.RoomEvent{
				RoomId: roomID,
			}

			switch rawEvent["type"] {
			case "USER_JOINED":
				roomEvent.Type = roompb.RoomEventType_ROOM_EVENT_TYPE_USER_JOINED
				
                userID, _ := rawEvent["user_id"].(string)
				username, _ := rawEvent["username"].(string)
				displayName, _ := rawEvent["display_name"].(string)
				avatarURL, _ := rawEvent["avatar_url"].(string)

				roomEvent.JoinedMember = &roompb.JoinedMember{
					UserId:      userID,
					Username:    username,
					DisplayName: displayName,
					AvatarUrl:   avatarURL,
				}

			case "USER_LEFT":
				roomEvent.Type = roompb.RoomEventType_ROOM_EVENT_TYPE_USER_LEFT

				userID, _ := rawEvent["user_id"].(string)
				roomEvent.LeftMember = &roompb.LeftMember{
					UserId: userID,
				}
			case "LIVE_CHAT_MESSAGE":
				roomEvent.Type = roompb.RoomEventType_ROOM_EVENT_TYPE_LIVE_CHAT_MESSAGE

				userID, _ := rawEvent["user_id"].(string)
				text, _ := rawEvent["text"].(string)

				roomEvent.LiveChatMessage = &roompb.LiveChatMessage{
					UserId: userID,
					Text:   text,
				}

			case "ROOM_ENDED":
				roomEvent.Type = roompb.RoomEventType_ROOM_EVENT_TYPE_ROOM_ENDED
			}

			if err := stream.Send(roomEvent); err != nil {
				return err
			}
		}
	}
}

func (h *GrpcHandler) ListRooms(ctx context.Context, req *roompb.ListRoomsRequest) (*roompb.ListRoomsResponse, error) {
	limit := int64(req.PageSize)
	if limit <= 0 {
		limit = 20
	} else if limit > 50 {
		limit = 50
	}

	offset := int64(0)
	if req.PageToken != "" {
		parsed, err := strconv.ParseInt(req.PageToken, 10, 64)
		if err == nil {
			offset = parsed
		}
	}

	summaries, nextOffset, err := h.service.ListActiveRooms(ctx, limit, offset)
	if err != nil {
		return nil, h.mapError(err)
	}

	nextToken := ""
	if nextOffset > 0 {
		nextToken = strconv.FormatInt(nextOffset, 10)
	}

	return &roompb.ListRoomsResponse{
		Rooms:         mapRoomSummariesToProto(summaries),
		NextPageToken: nextToken,
	}, nil
}

func (h *GrpcHandler) GetListeners(ctx context.Context, req *roompb.GetListenersRequest) (*roompb.GetListenersResponse, error) {
	listeners, err := h.service.GetListeners(ctx, req.RoomId)
	if err != nil {
		return nil, h.mapError(err)
	}

	totalListeners := int32(len(listeners))

	var pbListeners []*roompb.JoinedMember
	for _, l := range listeners {
		pbListeners = append(pbListeners, &roompb.JoinedMember{
			UserId:      l.UserID,
			Username:    l.Username,
			DisplayName: l.DisplayName,
			AvatarUrl:   l.AvatarURL,
		})
	}

	return &roompb.GetListenersResponse{
		Members:        pbListeners,
		TotalListeners: totalListeners,
	}, nil
}

func (h *GrpcHandler) SendLiveChatMessage(ctx context.Context, req *roompb.SendLiveChatMessageRequest) (*emptypb.Empty, error) {
	userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, h.mapError(err)
    }
	
	err = h.service.SendLiveChatMessage(ctx, req.RoomId, userID, req.Text)
	if err != nil {
		return nil, h.mapError(err)
	}
	return &emptypb.Empty{}, nil
}

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
