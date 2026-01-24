package routes

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb "github.com/dis70rt/bluppi-backend/internals/gen"
)

func Setup(server *grpc.Server, h *Handlers) {
	
	// Health Check (Standard Protocol)
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(server)
	
	// Users
	if h.UserHandler != nil {
		pb.RegisterUserServiceServer(server, h.UserHandler)
	}

	// Chat
	/* if h.ChatHandler != nil {
		pb.RegisterChatServiceServer(server, h.ChatHandler)
	} 
	*/
}