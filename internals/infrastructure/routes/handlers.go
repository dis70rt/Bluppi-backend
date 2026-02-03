package routes

import (
	"github.com/dis70rt/bluppi-backend/internals/music"
	"github.com/dis70rt/bluppi-backend/internals/users"
	"github.com/jmoiron/sqlx"
)

type Handlers struct {
	UserHandler *users.GrpcHandler
	TrackHandler *music.GrpcHandler
	// ChatHandler *chat.GrpcHandler
}

func BuildHandlers(db *sqlx.DB) *Handlers {
	// --- Users Module ---
	userRepo := users.NewRepository(db)
	userService := users.NewService(userRepo)
	userHandler := users.NewGrpcHandler(userService)

	// --- Tracks Modules ---
	trackRepo := music.NewRepository(db)
	trackService := music.NewService(trackRepo)
	trackHandler := music.NewGrpcHandler(trackService)

	// --- Future Modules ---
	// chatRepo := chat.NewRepository(db)

	return &Handlers{
		UserHandler: userHandler,
		TrackHandler: trackHandler,
		// ChatHandler: chatHandler,
	}
}