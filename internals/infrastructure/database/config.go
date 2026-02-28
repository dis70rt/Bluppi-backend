package database

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string

	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️  No .env file found, relying on System Env Vars")
	}

	return Config{
		Host:     mustEnv("DB_HOST", "localhost"),
		Port:     mustInt("DB_PORT", 5432),
		User:     mustEnv("DB_USER", "postgres"),
		Password: mustEnv("DB_PASSWORD", "password"),
		Name:     mustEnv("DB_NAME", "bluppi_db"),
		SSLMode:  mustEnv("DB_SSLMODE", "disable"),

		MaxConns:        int32(mustInt("DB_MAX_CONNS", 10)),
		MinConns:        int32(mustInt("DB_MIN_CONNS", 2)),
		MaxConnLifetime: mustDuration("DB_MAX_CONN_LIFETIME", time.Hour),
		MaxConnIdleTime: mustDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
	}
}

func LoadRedisConfig() RedisConfig {
    host := getEnv("REDIS_HOST", "localhost")
    port := getEnv("REDIS_PORT", "6379")
    
    dbInt, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

    return RedisConfig{
        Addr:     fmt.Sprintf("%s:%s", host, port),
        Password: getEnv("REDIS_PASSWORD", ""),
        DB:       dbInt,
    }
}

func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return fallback
}