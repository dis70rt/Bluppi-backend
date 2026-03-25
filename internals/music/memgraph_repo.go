package music

import (
    "context"
    "time"

    "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type GraphRepository struct {
    driver neo4j.DriverWithContext
}

func NewGraphRepository(driver neo4j.DriverWithContext) *GraphRepository {
    return &GraphRepository{driver: driver}
}

// LogListen records a user listening to a track, updating their current track and history.
func (g *GraphRepository) LogListen(ctx context.Context, userID, trackID string) error {
    session := g.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
    defer session.Close(ctx)

    _, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        query := `
            MERGE (u:User {id: $user_id})
            MERGE (t:Track {id: $track_id})
            MERGE (u)-[r:LISTENS_TO]->(t)
            SET r.last_played = $now, u.current_track_id = $track_id
        `
        return tx.Run(ctx, query, map[string]any{
            "user_id":  userID,
            "track_id": trackID,
            "now":      time.Now().Unix(),
        })
    })
    return err
}

// LikeTrack records a user liking a track in the graph.
func (g *GraphRepository) LikeTrack(ctx context.Context, userID, trackID string) error {
    session := g.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
    defer session.Close(ctx)

    _, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        query := `
            MERGE (u:User {id: $user_id})
            MERGE (t:Track {id: $track_id})
            MERGE (u)-[:LIKES]->(t)
        `
        return tx.Run(ctx, query, map[string]any{
            "user_id":  userID,
            "track_id": trackID,
        })
    })
    return err
}

// UnlikeTrack removes the like relationship between a user and a track.
func (g *GraphRepository) UnlikeTrack(ctx context.Context, userID, trackID string) error {
    session := g.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
    defer session.Close(ctx)

    _, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        query := `
            MATCH (u:User {id: $user_id})-[r:LIKES]->(t:Track {id: $track_id})
            DELETE r
        `
        return tx.Run(ctx, query, map[string]any{
            "user_id":  userID,
            "track_id": trackID,
        })
    })
    return err
}