package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/dis70rt/bluppi-backend/internals/gateway"
	pb "github.com/dis70rt/bluppi-backend/internals/gen/presences"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/database"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/firebase"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/middlewares"

	// "github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// func init() {
//     if err := godotenv.Load(); err != nil {
//         log.Println("ℹ️  No .env file found, relying on System Env Vars")
//     }
// }

func main() {
	middlewares.InitLogger()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	redisCfg := database.LoadRedisConfig()
	redisWrapper, err := database.NewRedis(redisCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisWrapper.Close()
	log.Info().Msg("Redis connected successfully")

	authClient, err := firebase.InitAuth()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Firebase Auth")
	}
	log.Info().Msg("Firebase Auth initialized")

	certFile := getEnv("TLS_CERT_FILE", "certs/server.crt")
	keyFile := getEnv("TLS_KEY_FILE", "certs/server.key")

	clientCreds, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load client TLS credentials")
	}

	// Server Credentials (used to secure the GATEWAY itself)
	serverCreds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load server TLS credentials")
	}

	// The internal gRPC API defaults to localhost:50051
	presenceServiceAddr := getEnv("PRESENCE_INTERNAL_URL", "localhost:50051")
	internalConn, err := grpc.NewClient(presenceServiceAddr, grpc.WithTransportCredentials(clientCreds))
	if err != nil {
		log.Fatal().Err(err).Msg("did not connect to internal presence service")
	}
	defer internalConn.Close()
	internalPresenceClient := pb.NewInternalPresenceClient(internalConn)

	// 4. Initialize Gateway Components
	connManager := gateway.NewConnectionManager()
	eventsListener := gateway.NewEventsListener(redisWrapper.Client, connManager)
	gatewayServer := gateway.NewServer(connManager, internalPresenceClient)

	go func() {
		log.Info().Msg("Starting Redis PubSub Events Listener...")
		eventsListener.Start(ctx)
	}()

	port := getEnv("GATEWAY_PORT", ":50050")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to listen on port %s", port)
	}

	kaep := keepalive.EnforcementPolicy{
		MinTime:             30 * time.Second,
		PermitWithoutStream: true,
	}

	kasp := keepalive.ServerParameters{
		MaxConnectionIdle: 5 * time.Minute,
		Time:              45 * time.Second,
		Timeout:           15 * time.Second,
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(serverCreds),
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
		grpc.ChainUnaryInterceptor(
			middlewares.RecoveryInterceptor(),
			middlewares.LoggingInterceptor(),
			middlewares.UnaryAuthInterceptor(authClient),
		),
		grpc.StreamInterceptor(middlewares.StreamAuthInterceptor(authClient)),
	)

	pb.RegisterPresenceGatewayServer(grpcServer, gatewayServer)

	go func() {
		log.Info().Msgf("Presence Gateway gRPC Server is running on localhost%s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("failed to serve gRPC")
		}
	}()

	<-ctx.Done()

	log.Info().Msg("Shutting down gateway server...")
	grpcServer.GracefulStop()
	log.Info().Msg("Gateway server safely stopped.")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
