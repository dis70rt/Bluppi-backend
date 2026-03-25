package activity

import (
    "context"
    "encoding/json"
    "log"

    "github.com/redis/go-redis/v9"
)

type PresenceEvent struct {
    UserID   string `json:"user_id"`
    Status   string `json:"status"`
    LastSeen int64  `json:"last_seen"`
}

type Consumer struct {
    redisClient *redis.Client
    graphRepo   *GraphRepository
}

func NewConsumer(redisClient *redis.Client, graphRepo *GraphRepository) *Consumer {
    return &Consumer{redisClient: redisClient, graphRepo: graphRepo}
}

func (c *Consumer) Start(ctx context.Context) {
    pubsub := c.redisClient.Subscribe(ctx, "system:presence_events")
    defer pubsub.Close()
    ch := pubsub.Channel()

    for {
        select {
        case <-ctx.Done():
            return
        case msg := <-ch:
            var event PresenceEvent
            if err := json.Unmarshal([]byte(msg.Payload), &event); err == nil {
                err := c.graphRepo.UpdateUserPresence(ctx, event.UserID, event.Status, event.LastSeen)
                if err != nil {
                    log.Printf("Activity Graph failed to update presence for %s: %v", event.UserID, err)
                }
            }
        }
    }
}