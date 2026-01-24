package database

import "time"

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
	return Config{
		Host:     mustEnv("DB_HOST", "localhost"),
		Port:     mustInt("DB_PORT", 5432),
		User:     mustEnv("DB_USER", ""),
		Password: mustEnv("DB_PASSWORD", ""),
		Name:     mustEnv("DB_NAME", ""),
		SSLMode:  mustEnv("DB_SSLMODE", "disable"),

		MaxConns:        int32(mustInt("DB_MAX_CONNS", 10)),
		MinConns:        int32(mustInt("DB_MIN_CONNS", 2)),
		MaxConnLifetime: mustDuration("DB_MAX_CONN_LIFETIME", time.Hour),
		MaxConnIdleTime: mustDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
	}
}
