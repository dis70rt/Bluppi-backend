package routes

import (
	"os"

	"github.com/dis70rt/bluppi-backend/internals/infrastructure/database"
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
	solr := database.NewSolrClient(os.Getenv("SOLR_URL"))
	trackRepo := music.NewRepository(db, solr)
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