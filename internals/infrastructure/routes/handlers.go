package routes

import (
	"context"
	"os"

	"firebase.google.com/go/v4/messaging"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/database"
	eventbus "github.com/dis70rt/bluppi-backend/internals/infrastructure/eventBus"
	"github.com/dis70rt/bluppi-backend/internals/music"
	"github.com/dis70rt/bluppi-backend/internals/notifications"
	"github.com/dis70rt/bluppi-backend/internals/party"
	"github.com/dis70rt/bluppi-backend/internals/playback"
	"github.com/dis70rt/bluppi-backend/internals/presence"
	"github.com/dis70rt/bluppi-backend/internals/users"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Handlers struct {
	UserHandler *users.GrpcHandler
	TrackHandler *music.GrpcHandler
	PartyHandler *party.GrpcHandler
	RoomHandler *party.GrpcHandler
	PlaybackHandler *playback.PlaybackHandler
	NotifHandler *notifications.GrpcHandler
	PresenceHandler *presence.Service
	// ChatHandler *chat.GrpcHandler
}

func BuildHandlers(ctx context.Context, db *sqlx.DB, redis *redis.Client, fcmClient *messaging.Client) *Handlers {
	globalBus := eventbus.NewRedisEventBus(redis)
	
	// --- Users Module ---
	userRepo := users.NewRepository(db)
	userService := users.NewService(userRepo, globalBus)
	userHandler := users.NewGrpcHandler(userService)

	// --- Tracks Modules ---
	solr := database.NewSolrClient(os.Getenv("SOLR_URL"))
	trackRepo := music.NewRepository(db, solr)
	trackService := music.NewService(trackRepo)
	trackHandler := music.NewGrpcHandler(trackService)

	// --- Party Module ---
	partyRepo := party.NewRepository(db)
	partyRedis := party.NewRedisRepository(redis)
	partyService := party.NewService(partyRepo, partyRedis, globalBus)
	partyHandler := party.NewGrpcHandler(partyService)
	roomHandler := party.NewGrpcHandler(partyService)

	partyReaper := party.NewReaper(partyRepo, partyRedis);
	go partyReaper.Start(ctx)

	roomManager := playback.NewRoomManager()
	playbackHandler := playback.NewPlaybackHandler(roomManager)

	notifRepo := notifications.NewRepository(db)
	notifService := notifications.NewService(notifRepo)
	notifHandler := notifications.NewGrpcHandler(notifService)

	fcmSender := notifications.NewFCMSender(fcmClient)
	notifConsumer := notifications.NewConsumer(notifService, globalBus, fcmSender)
	go notifConsumer.Start(ctx)

	presenceRepo := presence.NewRepository(redis)
	presenceService := presence.NewService(presenceRepo, redis)
	presenceReaper := presence.NewReaper(presenceRepo, redis)
	go presenceReaper.Start(ctx)

	// --- Future Modules ---
	// chatRepo := chat.NewRepository(db)

	return &Handlers{
		UserHandler: userHandler,
		TrackHandler: trackHandler,
		PartyHandler: partyHandler,
		RoomHandler: roomHandler,
		PlaybackHandler: playbackHandler,
		NotifHandler: notifHandler,
		PresenceHandler: presenceService,
		// ChatHandler: chatHandler,
	}
}