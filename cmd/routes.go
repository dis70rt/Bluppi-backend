package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	// pb "github.com/dis70rt/bluppi-backend/internals/gen"
	// "github.com/dis70rt/bluppi-backend/internals/users"
)

func RegisterGrpcRoutes(
	server *grpc.Server,
	// userHandler *users.GrpcHandler,
) {
	// 1. Register Health Check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(server)

	// if userHandler != nil {
	// 	pb.RegisterUserServiceServer(server, userHandler)
	// }
	
	// pb.RegisterChatServiceServer(server, chatHandler)
}