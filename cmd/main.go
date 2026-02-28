package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"github.com/dis70rt/bluppi-backend/internals/infrastructure/database"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/routes"
)

func main() {
	cfg := database.LoadConfig()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbWrapper, err := database.New(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer dbWrapper.Close()
	log.Println("✅ Database connected successfully")

	redisCfg := database.LoadRedisConfig()
	redisWrapper, err := database.NewRedis(redisCfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	defer redisWrapper.Close()
	appHandlers := routes.BuildHandlers(ctx, dbWrapper.DB, redisWrapper.Client)

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
