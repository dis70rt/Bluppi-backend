package activity

import (
    "context"

    pb "github.com/dis70rt/bluppi-backend/internals/gen/friends_activity"
    "github.com/dis70rt/bluppi-backend/internals/infrastructure/middlewares"
)

type GrpcHandler struct {
    pb.UnimplementedFriendsActivityServiceServer
    service *Service
}

func NewGrpcHandler(s *Service) *GrpcHandler {
    return &GrpcHandler{service: s}
}

func (h *GrpcHandler) GetFriendsFeed(ctx context.Context, req *pb.GetFriendsFeedRequest) (*pb.GetFriendsFeedResponse, error) {
    userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, err
    }

    activities, err := h.service.GetFriendsFeed(ctx, userID, req.Limit, req.Offset)
    if err != nil {
        return nil, mapError(err)
    }

    pbActivities := make([]*pb.FriendActivity, len(activities))
    for i, a := range activities {
        temp := a
        pbActivities[i] = h.mapActivityToProto(&temp)
    }

    return &pb.GetFriendsFeedResponse{
        Activities: pbActivities,
    }, nil
}