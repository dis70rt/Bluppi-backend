package routes

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb_users "github.com/dis70rt/bluppi-backend/internals/gen/users"
	pb_tracks "github.com/dis70rt/bluppi-backend/internals/gen/tracks"
	pb_party "github.com/dis70rt/bluppi-backend/internals/gen/party"
	pb_room "github.com/dis70rt/bluppi-backend/internals/gen/rooms"
	pb_playback "github.com/dis70rt/bluppi-backend/internals/gen/playback"
	pb_notif "github.com/dis70rt/bluppi-backend/internals/gen/notifications"
	pb_presence "github.com/dis70rt/bluppi-backend/internals/gen/presences"
)

func Setup(server *grpc.Server, h *Handlers) {
	
	// Health Check (Standard Protocol)
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(server)
	
	if h.UserHandler != nil {
		pb_users.RegisterUserServiceServer(server, h.UserHandler)
	}

	if h.TrackHandler != nil {
		pb_tracks.RegisterTrackServiceServer(server, h.TrackHandler)
	}

	if h.PartyHandler != nil {
		pb_party.RegisterSyncServiceServer(server, h.PartyHandler)
	}

	if h.RoomHandler != nil {
		pb_room.RegisterRoomServiceServer(server, h.RoomHandler)
	}

	if h.PlaybackHandler != nil {
		pb_playback.RegisterPlaybackServiceServer(server, h.PlaybackHandler)
	}

	if h.NotifHandler != nil {
		pb_notif.RegisterNotificationServiceServer(server, h.NotifHandler)
	}

	if h.PresenceHandler != nil {
        pb_presence.RegisterInternalPresenceServer(server, h.PresenceHandler)
    }
	
	/* if h.ChatHandler != nil {
		pb.RegisterChatServiceServer(server, h.ChatHandler)
	} 
	*/
}