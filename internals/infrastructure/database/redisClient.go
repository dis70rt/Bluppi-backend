package database

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisConfig struct {
    Addr     string
    Password string
    DB       int
}

type RedisDB struct {
    Client *redis.Client
}

func NewRedis(cfg RedisConfig) (*RedisDB, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     cfg.Addr,
        Password: cfg.Password,
        DB:       cfg.DB,
    })

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("connect redis: %w", err)
    }

    return &RedisDB{Client: client}, nil
}

func (r *RedisDB) Close() error {
    return r.Client.Close()
}

func Cached[T any](
    ctx context.Context,
    rdb *RedisDB,
    key string,
    ttl time.Duration,
    fetch func() (T, error),
) (T, error) {
    var result T

    val, err := rdb.Client.Get(ctx, key).Result()
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
        rdb.Client.Set(ctx, key, data, ttl)
    }

    return result, nil
}