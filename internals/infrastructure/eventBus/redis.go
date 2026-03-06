package eventbus

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

const streamName = "bluppi:global:events"

type RedisEventBus struct {
	client *redis.Client
}

func NewRedisEventBus(client *redis.Client) *RedisEventBus {
	return &RedisEventBus{
		client: client,
	}
}

func (r *RedisEventBus) Publish(ctx context.Context, topic EventType, payload proto.Message) error {
	data, err := proto.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf payload: %w", err)
	}

	args := &redis.XAddArgs{
		Stream: streamName,
		MaxLen: 10000,
		Approx: true,
		Values: map[string]interface{}{
			"type":    string(topic),
			"payload": data,
		},
	}

	_, err = r.client.XAdd(ctx, args).Result()
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

func (r *RedisEventBus) Consume(ctx context.Context, group string, consumerName string, topics []EventType, handler EventHandler) error {
	err := r.client.XGroupCreateMkStream(ctx, streamName, group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	topicMap := make(map[EventType]bool)
	for _, t := range topics {
		topicMap[t] = true
	}

	go r.startConsuming(ctx, group, consumerName, topicMap, handler)
	return nil
}

func (r *RedisEventBus) startConsuming(ctx context.Context, group string, consumer string, topicMap map[EventType]bool, handler EventHandler) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			streams, err := r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    group,
				Consumer: consumer,
				Streams:  []string{streamName, ">"},
				Count:    10,
				Block:    2 * time.Second,
			}).Result()

			if err != nil {
				if err != redis.Nil {
					log.Printf("redis bus error: %v", err)
					time.Sleep(1 * time.Second)
				}
				continue
			}

			for _, stream := range streams {
				for _, msg := range stream.Messages {
					eventTypeStr, ok1 := msg.Values["type"].(string)

					var payloadBytes []byte
					switch v := msg.Values["payload"].(type) {
					case string:
						payloadBytes = []byte(v)
					case []byte:
						payloadBytes = v
					}

					if ok1 && len(payloadBytes) > 0 {
						eventType := EventType(eventTypeStr)

						if topicMap[eventType] {
							event := Event{
								Type:    eventType,
								Payload: payloadBytes,
							}

							if err := handler(ctx, event); err == nil {
								r.client.XAck(ctx, streamName, group, msg.ID)
							} else {
								log.Printf("failed to process event %s: %v", msg.ID, err)
							}
						} else {
							r.client.XAck(ctx, streamName, group, msg.ID)
						}
					}
				}
			}
		}
	}
}

func (r *RedisEventBus) Close() error {
	return r.client.Close()
}