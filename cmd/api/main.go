package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"


	"github.com/dis70rt/bluppi-backend/internals/infrastructure/database"
	firebase "github.com/dis70rt/bluppi-backend/internals/infrastructure/firebase"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/middlewares"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/routes"
)

func main() {
	middlewares.InitLogger()

	cfg := database.LoadConfig()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbWrapper, err := database.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer dbWrapper.Close()
	log.Info().Msg("PostgreSQL connected successfully")

	redisCfg := database.LoadRedisConfig()
	redisWrapper, err := database.NewRedis(redisCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisWrapper.Close()
	log.Info().Msg("Redis connected successfully")

	authClient, err := firebase.InitAuth()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Firebase")
	}
	log.Info().Msg("Firebase Auth initialized")

	fcmClient, err := firebase.InitFCM()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Firebase FCM")
	}
	log.Info().Msg("Firebase FCM initialized")

	mgCfg := database.MemgraphConfig{
        URI:      getEnv("MEMGRAPH_URI", "bolt://localhost:7687"),
        Username: getEnv("MEMGRAPH_USER", ""),
        Password: getEnv("MEMGRAPH_PASS", ""),
    }
    mgWrapper, err := database.NewMemgraph(mgCfg)
    if err != nil {
        log.Fatal().Err(err).Msg("Failed to connect to Memgraph")
    }
    defer mgWrapper.Close(ctx)
    log.Info().Msg("Memgraph connected successfully")

	appHandlers := routes.BuildHandlers(ctx, dbWrapper.DB, redisWrapper.Client, mgWrapper.Driver, fcmClient)

	port := getEnv("PORT", ":50051")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to listen on %s", port)
	}

	certFile := getEnv("TLS_CERT_FILE", "certs/server.crt")
    keyFile := getEnv("TLS_KEY_FILE", "certs/server.key")

	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
    if err != nil {
        log.Fatal().Err(err).Msg("Failed to load TLS credentials")
    }

	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			middlewares.RecoveryInterceptor(),
			middlewares.LoggingInterceptor(),
			middlewares.UnaryAuthInterceptor(authClient),
		),
		grpc.StreamInterceptor(middlewares.StreamAuthInterceptor(authClient)),
	)

	routes.Setup(grpcServer, appHandlers)

	go func() {
		log.Info().Msgf("🚀 gRPC Server is running on localhost%s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("Failed to serve")
		}
	}()

	<-ctx.Done()

	log.Info().Msg("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	log.Info().Msg("gRPC server safely stopped.")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
