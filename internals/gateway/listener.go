package gateway

import (
    "context"
    "encoding/json"
    "log"

    "github.com/redis/go-redis/v9"
)

type EventsListener struct {
    redisClient *redis.Client
    connManager *ConnectionManager
}

func NewEventsListener(redisClient *redis.Client, connManager *ConnectionManager) *EventsListener {
    return &EventsListener{
        redisClient: redisClient,
        connManager: connManager,
    }
}

func (el *EventsListener) Start(ctx context.Context) {
    pubsub := el.redisClient.Subscribe(ctx, "system:presence_events")
    defer pubsub.Close()

    for {
        select {
        case <-ctx.Done():
            return
        case msg := <-pubsub.Channel():
            var event PresenceEvent
            if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
                log.Printf("Failed to unmarshal presence event: %v", err)
                continue
            }

            // Fan out the event to all users who subscribed to this Target UserID
            el.connManager.PushEvent(event.UserID, event)
        }
    }
}