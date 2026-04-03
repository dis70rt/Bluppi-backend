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

func (g *GraphRepository) GetWeeklyDiscover(ctx context.Context, userID string, limit int, cutoffTime int64) ([]string, error) {
    session := g.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
    defer session.Close(ctx)

    result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        query := `
            MATCH (u:User {id: $user_id})
            
            OPTIONAL MATCH (u)-[:FOLLOWS|LIKES|LISTENS_TO*1..3]-(s_user:User)
            WHERE s_user <> u
            
            MATCH (s_user)-[r:LIKES|LISTENS_TO]->(t:Track)
            
            WHERE ($cutoff == 0 OR r.last_played >= $cutoff OR type(r) = 'LIKES')
            
            AND NOT (u)-[:LIKES|LISTENS_TO]->(t)
            
            WITH t, count(s_user) as social_score
            ORDER BY social_score DESC
            LIMIT $limit
            
            RETURN t.id AS track_id
        `
        records, err := tx.Run(ctx, query, map[string]any{
            "user_id": userID,
            "limit":   limit,
            "cutoff":  cutoffTime,
        })
        if err != nil {
            return nil, err
        }

        var trackIDs []string
        for records.Next(ctx) {
            record := records.Record()
            trackID, ok := record.Get("track_id")
            if ok && trackID != nil {
                trackIDs = append(trackIDs, trackID.(string))
            }
        }
        return trackIDs, records.Err()
    })

    if err != nil {
        return nil, err
    }

    return result.([]string), nil
}