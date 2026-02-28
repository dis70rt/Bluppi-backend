package party

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

func activeRoomsSetKey() string {
	return "party:rooms:heartbeats"
}

func roomMembersKey(roomID string) string {
	return fmt.Sprintf("party:room:%s:members", roomID)
}

func roomChannelKey(roomID string) string {
    return fmt.Sprintf("party:room:%s:events", roomID)
}

// Reaper Room
func (r *RedisRepository) RecordHeartbeat(ctx context.Context, roomID string) error {
	score := float64(time.Now().Unix())
	err := r.client.ZAdd(ctx, activeRoomsSetKey(), redis.Z{
		Score:  score,
		Member: roomID,
	}).Err()

	return err
}

func (r *RedisRepository) GetDeadRooms(ctx context.Context, cutoff time.Time) ([]string, error) {
	maxScore := strconv.FormatInt(cutoff.Unix(), 10)

	rooms, err := r.client.ZRangeByScore(ctx, activeRoomsSetKey(), &redis.ZRangeBy{
		Min: "-inf",
		Max: maxScore,
	}).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	return rooms, nil
}

func (r *RedisRepository) GetRoomHost(ctx context.Context, roomID string) (string, error) {
	host, err := r.client.Get(ctx, fmt.Sprintf("party:room:%s:host", roomID)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", err
	}
	return host, nil
}

func (r *RedisRepository) SetRoomHost(ctx context.Context, roomID, hostUserID string) error {
    return r.client.Set(ctx, fmt.Sprintf("party:room:%s:host", roomID), hostUserID, 24*time.Hour).Err()
}

// Member state-management
func (r *RedisRepository) AddMember(ctx context.Context, roomID, userID string) error {
	return r.client.SAdd(ctx, roomMembersKey(roomID), userID).Err()
}

func (r *RedisRepository) RemoveMember(ctx context.Context, roomID, userID string) error {
	return r.client.SRem(ctx, roomMembersKey(roomID), userID).Err()
}

func (r *RedisRepository) GetMemberCount(ctx context.Context, roomID string) (int64, error) {
	count, err := r.client.SCard(ctx, roomMembersKey(roomID)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0, err
	}
	return count, nil
}

func (r *RedisRepository) CleanupRoom(ctx context.Context, roomID string) error {
	pipe := r.client.Pipeline()

	pipe.ZRem(ctx, activeRoomsSetKey(), roomID)
	pipe.Del(ctx, roomMembersKey(roomID))

	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisRepository) PublishRoomEvent(ctx context.Context, roomID string, event interface{}) error {
    payload, err := json.Marshal(event)
    if err != nil {
        return err
    }
    return r.client.Publish(ctx, roomChannelKey(roomID), payload).Err()
}

func (r *RedisRepository) SubscribeToRoom(ctx context.Context, roomID string) *redis.PubSub {
    return r.client.Subscribe(ctx, roomChannelKey(roomID))
}