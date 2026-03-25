package users

import (
    "context"

    "github.com/dis70rt/bluppi-backend/internals/utils"
    "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type GraphRepository struct {
    driver neo4j.DriverWithContext
}

func NewGraphRepository(driver neo4j.DriverWithContext) *GraphRepository {
    return &GraphRepository{driver: driver}
}

type FriendFeedItem struct {
    FriendID     string
    Status       string
    TrackID      *string
    SortPriority int64
}

// GetSortedFriendsFeed runs the "God Query" on Memgraph
func (g *GraphRepository) GetSortedFriendsFeed(ctx context.Context, userID string, limit, offset int32) ([]FriendFeedItem, error) {
    l, o := utils.SanitizePagination(limit, offset)

    session := g.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
    defer session.Close(ctx)

    res, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        query := `
            MATCH (me:User {id: $user_id})-[:FOLLOWS]->(friend:User)
            WITH friend,
                 CASE 
                     WHEN friend.status = 'online' AND friend.current_track_id IS NOT NULL 
                         THEN friend.current_track_id
                     WHEN friend.status = 'offline' OR friend.status IS NULL
                         THEN head([(friend)-[r:LISTENS_TO]->(t:Track) | {id: t.id, time: r.last_played} ORDER BY r.last_played DESC]).id
                     ELSE NULL 
                 END AS track_id,
                 CASE 
                     WHEN friend.status = 'online' AND friend.current_track_id IS NOT NULL THEN 1 
                     WHEN friend.status = 'online' AND friend.current_track_id IS NULL THEN 2     
                     ELSE 3                                                                       
                 END AS sort_priority
            ORDER BY sort_priority ASC, friend.id ASC
            SKIP $offset LIMIT $limit
            RETURN friend.id AS friend_id, friend.status AS status, track_id, sort_priority
        `

        result, err := tx.Run(ctx, query, map[string]any{
            "user_id": userID,
            "offset":  o,
            "limit":   l,
        })
        if err != nil {
            return nil, err
        }

        var feed []FriendFeedItem
        for result.Next(ctx) {
            record := result.Record()
            fID, _ := record.Get("friend_id")
            status, _ := record.Get("status")
            trackID, _ := record.Get("track_id")
            priority, _ := record.Get("sort_priority")

            var tID *string
            if trackID != nil && trackID.(string) != "" {
                str := trackID.(string)
                tID = &str
            }

            statusStr := "offline"
            if status != nil && status.(string) != "" {
                statusStr = status.(string)
            }

            feed = append(feed, FriendFeedItem{
                FriendID:     fID.(string),
                Status:       statusStr,
                TrackID:      tID,
                SortPriority: priority.(int64),
            })
        }
        return feed, result.Err()
    })

    if err != nil {
        return nil, err
    }
    return res.([]FriendFeedItem), nil
}

func (g *GraphRepository) Follow(ctx context.Context, followerID, followeeID string) error {
    session := g.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
    defer session.Close(ctx)

    _, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        query := `
            MERGE (u1:User {id: $follower_id})
            MERGE (u2:User {id: $followee_id})
            MERGE (u1)-[:FOLLOWS]->(u2)
        `
        return tx.Run(ctx, query, map[string]any{
            "follower_id": followerID,
            "followee_id": followeeID,
        })
    })

    return err
}

// Unfollow removes the directed FOLLOWS edge between two users in Memgraph if it exists.
func (g *GraphRepository) Unfollow(ctx context.Context, followerID, followeeID string) error {
    session := g.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
    defer session.Close(ctx)

    _, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        query := `
            MATCH (u1:User {id: $follower_id})-[r:FOLLOWS]->(u2:User {id: $followee_id})
            DELETE r
        `
        return tx.Run(ctx, query, map[string]any{
            "follower_id": followerID,
            "followee_id": followeeID,
        })
    })

    return err
}