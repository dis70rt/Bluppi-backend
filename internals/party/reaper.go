package party

import (
	"context"
	"log"
	"time"
)

type Reaper struct {
	repo      *Repository
	redisRepo *RedisRepository
	timeout   time.Duration
	interval  time.Duration
}

func NewReaper(repo *Repository, redisRepo *RedisRepository) *Reaper {
	return &Reaper{
		repo:      repo,
		redisRepo: redisRepo,
		timeout:   15 * time.Second,
		interval:  10 * time.Second,
	}
}

func (r *Reaper) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	log.Println("Starting Room Reaper worker...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down Room Reaper worker.")
			return
		case <-ticker.C:
			r.sweepDeadRooms(context.Background())
		}
	}
}

func (r *Reaper) sweepDeadRooms(ctx context.Context) {
	cutoff := time.Now().Add(-r.timeout)

	deadRooms, err := r.redisRepo.GetDeadRooms(ctx, cutoff)
	if err != nil {
		log.Printf("Reaper error fetching dead rooms: %v", err)
		return
	}

	for _, roomID := range deadRooms {
		log.Printf("Room %s heartbeat timed out. Marking as ENDED.", roomID)

		err := r.repo.EndRoom(ctx, roomID) 
		if err != nil {
			log.Printf("Failed to end room %s in PG: %v", roomID, err)
			continue
		}

		event := map[string]interface{}{
            "type":    "ROOM_ENDED",
            "room_id": roomID,
        }

		_ = r.redisRepo.PublishRoomEvent(ctx, roomID, event)
		_ = r.redisRepo.CleanupRoom(ctx, roomID)
	}
}