package presence

import (
    "context"
    "encoding/json"
    "log"
    "time"

    "github.com/dis70rt/bluppi-backend/internals/gateway"
    "github.com/redis/go-redis/v9"
)

type Reaper struct {
    repo        *Repository
    redisClient *redis.Client
}

func NewReaper(repo *Repository, redisClient *redis.Client) *Reaper {
    return &Reaper{
        repo:        repo,
        redisClient: redisClient,
    }
}

func (r *Reaper) Start(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            r.sweep(ctx)
        }
    }
}

func (r *Reaper) sweep(ctx context.Context) {
    // Timeout for unresponsive Redis
    sweepCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
    defer cancel()

    expiredUsers, err := r.repo.GetAndRemoveExpiredUsers(sweepCtx)
    if err != nil {
        log.Printf("Reaper failed to fetch expired users: %v", err)
        return
    }

    if len(expiredUsers) > 0 {
        now := time.Now().Unix()
        for _, userID := range expiredUsers {

            event := gateway.PresenceEvent{
                UserID:   userID,
                Status:   "offline",
                LastSeen: now,
            }

            payload, err := json.Marshal(event)
            if err == nil {
                r.redisClient.Publish(sweepCtx, "system:presence_events", payload)
            }
        }
    }
}