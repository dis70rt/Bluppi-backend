package routes

import (
	"github.com/dis70rt/bluppi-backend/internals/users"
	"github.com/jmoiron/sqlx"
)

type Handlers struct {
	UserHandler *users.GrpcHandler
	// ChatHandler *chat.GrpcHandler
}

func BuildHandlers(db *sqlx.DB) *Handlers {
	// --- Users Module ---
	userRepo := users.NewRepository(db)
	userService := users.NewService(userRepo)
	userHandler := users.NewGrpcHandler(userService)

	// --- Future Modules ---
	// chatRepo := chat.NewRepository(db)

	return &Handlers{
		UserHandler: userHandler,
		// ChatHandler: chatHandler,
	}
}