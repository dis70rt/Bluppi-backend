package gateway

import (
	"context"
	"log"
	"time"

	pb "github.com/dis70rt/bluppi-backend/internals/gen/presences"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/middlewares"
	"google.golang.org/grpc/metadata"
)

type Server struct {
    pb.UnimplementedPresenceGatewayServer
    connManager    *ConnectionManager
    internalClient pb.InternalPresenceClient
}

func NewServer(connManager *ConnectionManager, internalClient pb.InternalPresenceClient) *Server {
    return &Server{
        connManager:    connManager,
        internalClient: internalClient,
    }
}

func (s *Server) SubscribePresence(req *pb.SubscribeRequest, stream pb.PresenceGateway_SubscribePresenceServer) error {
    ctx := stream.Context()

	userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return err
    }

    // Track the specific connection and their specific targets
    conn, connID := s.connManager.AddConnection(userID, req.TargetUserIds)
    defer s.connManager.RemoveConnection(userID, connID)

    go s.keepAliveLoop(stream.Context(), userID)

    for {
        select {
        case <-stream.Context().Done():
            return stream.Context().Err()

        case event, ok := <-conn.Chan:
            if !ok {
                // The channel was closed gracefully by RemoveConnection
                return nil
            }

            grpcMessage := &pb.PresenceUpdate{
                UserId:   event.UserID,
                Status:   event.Status,
                LastSeen: event.LastSeen,
            }

            if err := stream.Send(grpcMessage); err != nil {
                return err
            }
        }
    }
}

func (s *Server) keepAliveLoop(ctx context.Context, userID string) {
    ticker := time.NewTicker(15 * time.Second)
    defer ticker.Stop()

    s.sendHeartbeat(userID)

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            s.sendHeartbeat(userID)
        }
    }
}

func (s *Server) sendHeartbeat(userID string) {
    // Add proper timeout to prevent heartbeat from hanging goroutines forever
    hbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    hbCtx = metadata.AppendToOutgoingContext(hbCtx, "x-mock-user-id", userID)

    _, err := s.internalClient.RecordHeartbeat(hbCtx, &pb.HeartBeatRequest{UserId: userID})
    if err != nil {
        // Log failures so we know if the presence service goes down
        log.Printf("Heartbeat failed for user %s: %v", userID, err)
    }
}