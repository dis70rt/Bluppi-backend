package tests

import (
    "fmt"
    "log"
    "os"
    "strconv"
    "testing"

    "github.com/jmoiron/sqlx"
    "github.com/joho/godotenv"
    _ "github.com/lib/pq"
)

type TestConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    Name     string
    SSLMode  string
}

func LoadTestConfig() TestConfig {
    if err := godotenv.Load("../../.env"); err != nil {
        log.Println("ℹ️  No .env file found, relying on System Env Vars")
    }

    return TestConfig{
        Host:     getEnv("TEST_DB_HOST", "localhost"),
        Port:     getEnvInt("TEST_DB_PORT", 5432),
        User:     getEnv("TEST_DB_USER", "postgres"),
        Password: getEnv("TEST_DB_PASSWORD", "password"),
        Name:     getEnv("TEST_DB_NAME", "bluppi_test"),
        SSLMode:  getEnv("TEST_DB_SSLMODE", "disable"),
    }
}

func (c TestConfig) DSN() string {
    return fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        c.Host,
        c.Port,
        c.User,
        c.Password,
        c.Name,
        c.SSLMode,
    )
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}

func getEnvInt(key string, fallback int) int {
    v := os.Getenv(key)
    if v == "" {
        return fallback
    }
    i, err := strconv.Atoi(v)
    if err != nil {
        return fallback
    }
    return i
}

func GetTestDB(t *testing.T) *sqlx.DB {
    t.Helper()

    cfg := LoadTestConfig()
    db, err := sqlx.Connect("postgres", cfg.DSN())
    if err != nil {
        t.Fatalf("failed to connect to test database: %v", err)
    }

    return db
}

func StringPtr(s string) *string {
    return &s
}