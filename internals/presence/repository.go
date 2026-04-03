package presence

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

const (
    presenceZSetKey = "presence:active_users"
    ttlSeconds      = 30
)

type Repository struct {
    redisClient *redis.Client
}

func NewRepository(redisClient *redis.Client) *Repository {
    return &Repository{
        redisClient: redisClient,
    }
}

func (r *Repository) RecordHeartbeat(ctx context.Context, userID string) (bool, error) {
    now := float64(time.Now().Unix())

    cutoff := float64(time.Now().Unix() - ttlSeconds)

    score, err := r.redisClient.ZScore(ctx, presenceZSetKey, userID).Result()
    var isNewConnection bool

    if err == redis.Nil {
        isNewConnection = true
    } else if err == nil && score < cutoff {
        isNewConnection = true
    } else if err != nil {
        return false, err
    }

    err = r.redisClient.ZAdd(ctx, presenceZSetKey, redis.Z{
        Score:  now,
        Member: userID,
    }).Err()

    return isNewConnection, err
}

func (r *Repository) GetAndRemoveExpiredUsers(ctx context.Context) ([]string, error) {
    cutoff := float64(time.Now().Unix() - ttlSeconds)
    maxScore := fmt.Sprintf("%f", cutoff)

    // Get all users with a score between -inf and (now - 30s) using the new ZRangeArgs
    expiredUsers, err := r.redisClient.ZRangeArgs(ctx, redis.ZRangeArgs{
        Key:     presenceZSetKey,
        ByScore: true,
        Start:   "-inf",
        Stop:    maxScore,
    }).Result()

    if err != nil {
        return nil, err
    }

    if len(expiredUsers) > 0 {
        members := make([]interface{}, len(expiredUsers))
        for i, u := range expiredUsers {
            members[i] = u
        }

        err = r.redisClient.ZRem(ctx, presenceZSetKey, members...).Err()
        if err != nil {
            return nil, err
        }
    }

    return expiredUsers, nil
}