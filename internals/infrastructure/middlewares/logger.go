package middlewares

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		latency := time.Since(start)
		if err != nil {
			log.Printf("[gRPC ERROR] %s | Latency: %v | Error: %v", info.FullMethod, latency, err)
		} else {
			log.Printf("[gRPC OK] %s | Latency: %v", info.FullMethod, latency)
		}

		return resp, err
	}
}