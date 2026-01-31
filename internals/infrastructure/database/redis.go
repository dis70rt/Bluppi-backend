package database

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"
)

func NewRedisClient(addr string, password string, db int) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })
}

func Cached[T any](
    ctx context.Context,
    client *redis.Client,
    key string,
    ttl time.Duration,
    fetch func() (T, error),
) (T, error) {
    var result T

    val, err := client.Get(ctx, key).Result()
    if err == nil {
        
        if err := json.Unmarshal([]byte(val), &result); err == nil {
            return result, nil
        }
    }

    result, err = fetch()
    if err != nil {
        return result, err
    }

    data, err := json.Marshal(result)
    if err == nil {
        // We set the cache with expiration.
        // Note: We ignore Redis Set errors here so that if Redis is down, 
        // the main application flow (returning data to user) is not interrupted.
        client.Set(ctx, key, data, ttl)
    }

    return result, nil
}