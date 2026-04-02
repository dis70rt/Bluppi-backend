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
	"google.golang.org/grpc/credentials"


	"github.com/dis70rt/bluppi-backend/internals/infrastructure/database"
	firebase "github.com/dis70rt/bluppi-backend/internals/infrastructure/firebase"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/middlewares"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/routes"
)

func main() {
	cfg := database.LoadConfig()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbWrapper, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbWrapper.Close()
	log.Println("PostgreSQL connected successfully")

	redisCfg := database.LoadRedisConfig()
	redisWrapper, err := database.NewRedis(redisCfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisWrapper.Close()
	log.Println("Redis connected successfully")

	authClient, err := firebase.InitAuth()
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}
	log.Println("Firebase Auth initialized")

	fcmClient, err := firebase.InitFCM()
	if err != nil {
		log.Fatalf("Failed to initialize Firebase FCM: %v", err)
	}
	log.Println("Firebase FCM initialized")

	mgCfg := database.MemgraphConfig{
        URI:      getEnv("MEMGRAPH_URI", "bolt://localhost:7687"),
        Username: getEnv("MEMGRAPH_USER", ""),
        Password: getEnv("MEMGRAPH_PASS", ""),
    }
    mgWrapper, err := database.NewMemgraph(mgCfg)
    if err != nil {
        log.Fatalf("Failed to connect to Memgraph: %v", err)
    }
    defer mgWrapper.Close(ctx)
    log.Println("Memgraph connected successfully")

	appHandlers := routes.BuildHandlers(ctx, dbWrapper.DB, redisWrapper.Client, mgWrapper.Driver, fcmClient)

	port := getEnv("PORT", ":50051")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", port, err)
	}

	certFile := getEnv("TLS_CERT_FILE", "certs/server.crt")
    keyFile := getEnv("TLS_KEY_FILE", "certs/server.key")

	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
    if err != nil {
        log.Fatalf("Failed to load TLS credentials: %v", err)
    }

	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(middlewares.UnaryAuthInterceptor(authClient)),
		grpc.StreamInterceptor(middlewares.StreamAuthInterceptor(authClient)),
	)

	routes.Setup(grpcServer, appHandlers)

	go func() {
		log.Printf("🚀 gRPC Server is running on localhost%s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	log.Println("gRPC server safely stopped.")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
