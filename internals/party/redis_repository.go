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

func lobbyLeaderboardKey() string {
	return "party:lobby:leaderboard"
}

func roomMetaKey(roomID string) string {
	return fmt.Sprintf("party:room:%s:meta", roomID)
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

func (r *RedisRepository) PublishToLobby(ctx context.Context, room *RoomSummary) error {
	pipe := r.client.Pipeline()
	
	pipe.HSet(ctx, roomMetaKey(room.RoomID), map[string]interface{}{
		"room_name":         room.RoomName,
		"host_user_id":      room.HostUserID,
		"host_display_name": room.HostDisplayName,
		"visibility":        string(room.Visibility),
	})
	
	pipe.ZAdd(ctx, lobbyLeaderboardKey(), redis.Z{
		Score:  1,
		Member: room.RoomID,
	})

	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisRepository) SetLobbyCount(ctx context.Context, roomID string, exactCount int64) error {
    return r.client.ZAdd(ctx, lobbyLeaderboardKey(), redis.Z{
        Score:  float64(exactCount),
        Member: roomID,
    }).Err()
}

func (r *RedisRepository) UpdateLobbyTrack(ctx context.Context, roomID, trackID, title, artist, artURL string) error {
	return r.client.HSet(ctx, roomMetaKey(roomID), map[string]interface{}{
		"track_id":     trackID,
		"track_title":  title,
		"track_artist": artist,
		"artwork_url":  artURL,
	}).Err()
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
	pipe.Del(ctx, fmt.Sprintf("party:room:%s:host", roomID))

	pipe.ZRem(ctx, lobbyLeaderboardKey(), roomID)
    pipe.Del(ctx, roomMetaKey(roomID))

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

func (r *RedisRepository) GetLobbyFeed(ctx context.Context, limit int64, offset int64) ([]RoomSummary, error) {
	zKeys, err := r.client.ZRevRangeWithScores(ctx, lobbyLeaderboardKey(), offset, offset+limit-1).Result()
	if err != nil {
		return nil, err
	}
	if len(zKeys) == 0 {
		return []RoomSummary{}, nil
	}

	pipe := r.client.Pipeline()
	var hashCmds []*redis.MapStringStringCmd

	for _, z := range zKeys {
		roomID := z.Member.(string)
		cmd := pipe.HGetAll(ctx, roomMetaKey(roomID))
		hashCmds = append(hashCmds, cmd)
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	var summaries []RoomSummary
	for i, cmd := range hashCmds {
		data := cmd.Val()
		if len(data) == 0 {
			continue
		}

		summaries = append(summaries, RoomSummary{
			RoomID:          zKeys[i].Member.(string),
			RoomName:        data["room_name"],
			HostUserID:      data["host_user_id"],
			HostDisplayName: data["host_display_name"],
			TrackID:         data["track_id"],
			TrackTitle:      data["track_title"],
			TrackArtist:     data["track_artist"],
			ArtworkURL:      data["artwork_url"],
			ListenerCount:   int32(zKeys[i].Score),
			IsLive:          true,
			Visibility:      RoomVisibility(data["visibility"]),
		})
	}

	return summaries, nil
}