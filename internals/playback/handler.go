package playback

import (
	"fmt"
	"io"
	"log"

	pb "github.com/dis70rt/bluppi-backend/internals/gen/playback"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/middlewares"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PlaybackHandler struct {
	pb.UnimplementedPlaybackServiceServer
	roomManager *RoomManager
}

func NewPlaybackHandler(rm *RoomManager) *PlaybackHandler {
	return &PlaybackHandler{
		roomManager: rm,
	}
}

func (h *PlaybackHandler) StreamSession(stream pb.PlaybackService_StreamSessionServer) error {
	firstCmd, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		log.Printf("Error reading initial command: %v", err)
		return err
	}

	roomID := firstCmd.RoomId
	
	userID, err := middlewares.GetUserID(stream.Context())
    if err != nil {
        return status.Error(codes.Unauthenticated, "user not authenticated")
    }

	if roomID == "" || userID == "" {
		return fmt.Errorf("initial command missing room_id or user_id")
	}

	room := h.roomManager.GetRoom(userID, roomID)

	client := &ClientSession{
		UserID: userID,
		Send:   make(chan *pb.ServerEvent, 100),
	}

	room.RegisterClient(client)

	defer func() {
		room.UnregisterClient(userID)
		close(client.Send)
	}()

	room.SendCurrentState(client)

	go func() {
		for event := range client.Send {
			if err := stream.Send(event); err != nil {
				log.Printf("Error pushing to client %s: %v", userID, err)
				return
			}
		}
	}()

	h.processCommand(room, userID, firstCmd)

	for {
		cmd, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Printf("Client %s disconnected cleanly", userID)
				break
			}
			log.Printf("Stream read error for client %s: %v", userID, err)
			break
		}

		h.processCommand(room, userID, cmd)
	}

	return nil
}

func (h *PlaybackHandler) processCommand(room *RoomState, userID string, cmd *pb.ClientCommand) {
	log.Printf("📥 [gRPC] Received command for room %s | Payload Type: %T", room.ID, cmd.Payload)
	switch payload := cmd.Payload.(type) {

	case *pb.ClientCommand_TrackChange:
		log.Printf("🎵 [gRPC] Track Change: %s - %s", payload.TrackChange.Title, payload.TrackChange.Artist)
		room.HandleTrackChange(payload.TrackChange)

	case *pb.ClientCommand_BufferReady:
		log.Printf("✅ [gRPC] User %s Buffer Ready (v%d)", userID, payload.BufferReady.Version)
		room.HandleBufferReady(userID, payload.BufferReady.Version)

	case *pb.ClientCommand_Play:
		log.Printf("▶️ [gRPC] Play Command")
		room.HandlePlay()

	case *pb.ClientCommand_Pause:
		log.Printf("⏸️ [gRPC] Pause Command")
		room.HandlePause()

	default:
		log.Printf("ℹ️ [gRPC] Empty/Initial Payload")
	}
}