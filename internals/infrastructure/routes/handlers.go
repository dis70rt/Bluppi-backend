package routes

import (
	"context"
	"os"

	"github.com/dis70rt/bluppi-backend/internals/infrastructure/database"
	"github.com/dis70rt/bluppi-backend/internals/music"
	"github.com/dis70rt/bluppi-backend/internals/party"
	"github.com/dis70rt/bluppi-backend/internals/users"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Handlers struct {
	UserHandler *users.GrpcHandler
	TrackHandler *music.GrpcHandler
	PartyHandler *party.GrpcHandler
	RoomHandler *party.GrpcHandler
	// ChatHandler *chat.GrpcHandler
}

func BuildHandlers(ctx context.Context, db *sqlx.DB, redis *redis.Client) *Handlers {
	// --- Users Module ---
	userRepo := users.NewRepository(db)
	userService := users.NewService(userRepo)
	userHandler := users.NewGrpcHandler(userService)

	// --- Tracks Modules ---
	solr := database.NewSolrClient(os.Getenv("SOLR_URL"))
	trackRepo := music.NewRepository(db, solr)
	trackService := music.NewService(trackRepo)
	trackHandler := music.NewGrpcHandler(trackService)

	// --- Party Module ---
	partyRepo := party.NewRepository(db)
	partyRedis := party.NewRedisRepository(redis)
	partyService := party.NewService(partyRepo, partyRedis)
	partyHandler := party.NewGrpcHandler(partyService)
	roomHandler := party.NewGrpcHandler(partyService)

	partyReaper := party.NewReaper(partyRepo, partyRedis);
	go partyReaper.Start(ctx)

	// --- Future Modules ---
	// chatRepo := chat.NewRepository(db)

	return &Handlers{
		UserHandler: userHandler,
		TrackHandler: trackHandler,
		PartyHandler: partyHandler,
		RoomHandler: roomHandler,
		// ChatHandler: chatHandler,
	}
}