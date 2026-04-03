package database

import (
    "context"
    "fmt"
    "time"

    "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type MemgraphConfig struct {
    URI      string
    Username string
    Password string
}

type MemgraphDB struct {
    Driver neo4j.DriverWithContext
}

func NewMemgraph(cfg MemgraphConfig) (*MemgraphDB, error) {
    driver, err := neo4j.NewDriverWithContext(
        cfg.URI,
        neo4j.BasicAuth(cfg.Username, cfg.Password, ""),
    )
    if err != nil {
        return nil, fmt.Errorf("connect memgraph: %w", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := driver.VerifyConnectivity(ctx); err != nil {
        driver.Close(ctx)
        return nil, fmt.Errorf("verify memgraph connectivity: %w", err)
    }

    return &MemgraphDB{Driver: driver}, nil
}

func (m *MemgraphDB) Close(ctx context.Context) error {
    return m.Driver.Close(ctx)
}