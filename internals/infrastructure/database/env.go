package database

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func mustEnv(key string, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if def != "" {
		return def
	}
	panic(fmt.Sprintf("missing required env var: %s", key))
}

func mustInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Sprintf("invalid int for %s: %v", key, err))
	}
	return i
}

func mustDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
    }

	d, err := time.ParseDuration(v)
	if err != nil {
		panic(fmt.Sprintf("invalid duration for %s: %v", key, err))
	}
	return d
}