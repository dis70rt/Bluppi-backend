package presence

import (
    "context"
    "encoding/json"
    "log"
    "time"

    "github.com/dis70rt/bluppi-backend/internals/gateway"
    pb "github.com/dis70rt/bluppi-backend/internals/gen/presences"
    "github.com/redis/go-redis/v9"
    "google.golang.org/protobuf/types/known/emptypb"
)

type Service struct {
    pb.UnimplementedInternalPresenceServer
    repo        *Repository
    redisClient *redis.Client
}

func NewService(repo *Repository, redisClient *redis.Client) *Service {
    return &Service{
        repo:        repo,
        redisClient: redisClient,
    }
}

func (s *Service) RecordHeartbeat(ctx context.Context, req *pb.HeartBeatRequest) (*emptypb.Empty, error) {
    isNewConnection, err := s.repo.RecordHeartbeat(ctx, req.UserId)
    if err != nil {
        log.Printf("Failed to record heartbeat for user %s: %v", req.UserId, err)
        return &emptypb.Empty{}, err
    }

    if isNewConnection {

        event := gateway.PresenceEvent{
            UserID:   req.UserId,
            Status:   "online",
            LastSeen: time.Now().Unix(),
        }
        
        payload, err := json.Marshal(event)
        if err == nil {
            s.redisClient.Publish(ctx, "system:presence_events", payload)
        }
    }

    return &emptypb.Empty{}, nil
}