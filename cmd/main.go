package main

import (
	"log"
	"net"
	"os"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"github.com/dis70rt/bluppi-backend/internals/infrastructure/database"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/routes"
)

func main() {
	cfg := database.LoadConfig()

	dbWrapper, err := database.New(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer dbWrapper.Close()

	log.Println("✅ Database connected successfully")
	appHandlers := routes.BuildHandlers(dbWrapper.DB)

	port := getEnv("PORT", ":50051")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("❌ Failed to listen on %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()

	routes.Setup(grpcServer, appHandlers)

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