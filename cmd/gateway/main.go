package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/dis70rt/bluppi-backend/internals/gateway"
	pb "github.com/dis70rt/bluppi-backend/internals/gen/presences"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/database"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/firebase"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/middlewares"

	// "github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// func init() {
//     if err := godotenv.Load(); err != nil {
//         log.Println("ℹ️  No .env file found, relying on System Env Vars")
//     }
// }

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    redisCfg := database.LoadRedisConfig()
    redisWrapper, err := database.NewRedis(redisCfg)
    if err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }
    defer redisWrapper.Close()
    log.Println("Redis connected successfully")

    authClient, err := firebase.InitAuth()
    if err != nil {
        log.Fatalf("Failed to initialize Firebase Auth: %v", err)
    }
    log.Println("Firebase Auth initialized")

    certFile := getEnv("TLS_CERT_FILE", "certs/server.crt")
    keyFile := getEnv("TLS_KEY_FILE", "certs/server.key")

    clientCreds, err := credentials.NewClientTLSFromFile(certFile, "")
    if err != nil {
        log.Fatalf("Failed to load client TLS credentials: %v", err)
    }

    // Server Credentials (used to secure the GATEWAY itself)
    serverCreds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
    if err != nil {
        log.Fatalf("Failed to load server TLS credentials: %v", err)
    }

    // The internal gRPC API defaults to localhost:50051
    presenceServiceAddr := getEnv("PRESENCE_INTERNAL_URL", "localhost:50051")
    internalConn, err := grpc.NewClient(presenceServiceAddr, grpc.WithTransportCredentials(clientCreds))
    if err != nil {
        log.Fatalf("did not connect to internal presence service: %v", err)
    }
    defer internalConn.Close()
    internalPresenceClient := pb.NewInternalPresenceClient(internalConn)

    // 4. Initialize Gateway Components
    connManager := gateway.NewConnectionManager()
    eventsListener := gateway.NewEventsListener(redisWrapper.Client, connManager)
    gatewayServer := gateway.NewServer(connManager, internalPresenceClient)

    go func() {
        log.Println("Starting Redis PubSub Events Listener...")
        eventsListener.Start(ctx)
    }()

    port := getEnv("GATEWAY_PORT", ":50050")
    lis, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatalf("failed to listen on port %s: %v", port, err)
    }

    grpcServer := grpc.NewServer(
        grpc.Creds(serverCreds),
        grpc.UnaryInterceptor(middlewares.UnaryAuthInterceptor(authClient)),
        grpc.StreamInterceptor(middlewares.StreamAuthInterceptor(authClient)),
    )
    
    pb.RegisterPresenceGatewayServer(grpcServer, gatewayServer)

    go func() {
        log.Printf("Presence Gateway gRPC Server is running on localhost%s", port)
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("failed to serve gRPC: %v", err)
        }
    }()

    <-ctx.Done()
    
    log.Println("Shutting down gateway server...")
    grpcServer.GracefulStop()
    log.Println("Gateway server safely stopped.")
}

func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return fallback
}