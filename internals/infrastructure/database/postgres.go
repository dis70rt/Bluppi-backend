package database

import (
    "fmt"
    "time"

    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

type Database struct {
    DB *sqlx.DB
}

func New(cfg Config) (*Database, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host,
        cfg.Port,
        cfg.User,
        cfg.Password,
        cfg.Name,
        cfg.SSLMode,
    )

    db, err := sqlx.Connect("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("connect postgres: %w", err)
    }

    db.SetConnMaxIdleTime(5 * time.Minute)

    return &Database{DB: db}, nil
}

func (d *Database) Close() error {
    return d.DB.Close()
}