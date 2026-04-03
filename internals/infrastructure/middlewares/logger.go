package middlewares

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// ErrorCallerHook adds caller details to the log event only for errors/fatals
type ErrorCallerHook struct{}

func (h ErrorCallerHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level == zerolog.ErrorLevel || level == zerolog.FatalLevel || level == zerolog.PanicLevel {
		e.Caller(zerolog.CallerSkipFrameCount + 1)
	}
}

func InitLogger() {
	// Standardize time format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05.000",
		NoColor:    false,
	}

	logger := zerolog.New(consoleWriter).
		With().
		Timestamp().
		Logger().
		Hook(ErrorCallerHook{})

	// Replace the global logger
	log.Logger = logger
}

func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		latency := time.Since(start)
		if err != nil {
			log.Error().
				Str("method", info.FullMethod).
				Dur("latency", latency).
				Err(err).
				Msg("gRPC ERROR")
		} else {
			log.Info().
				Str("method", info.FullMethod).
				Dur("latency", latency).
				Msg("gRPC OK")
		}

		return resp, err
	}
}