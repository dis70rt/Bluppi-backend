package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	_ "github.com/lib/pq"

	// "github.com/dis70rt/bluppi-backend/internals/infrastructure/database"
	// "github.com/dis70rt/bluppi-backend/internals/users"
)

func main() {
	port := getEnv("PORT", ":50051")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("❌ Failed to listen on %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	RegisterGrpcRoutes(grpcServer /*, userHandler */)

	log.Printf("🚀 gRPC Server is running on localhost%s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("❌ Failed to serve: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}