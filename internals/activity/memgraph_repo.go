package activity

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

func (g *GraphRepository) UpdateUserPresence(ctx context.Context, userID, status string, lastSeen int64) error {
    session := g.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
    defer session.Close(ctx)

    _, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        query := `
            MERGE (u:User {id: $user_id})
            SET u.status = $status,
                u.last_seen = $last_seen,
                u.current_track_id = CASE WHEN $status = 'offline' THEN NULL ELSE u.current_track_id END
        `
        return tx.Run(ctx, query, map[string]any{
            "user_id":   userID,
            "status":    status,
            "last_seen": lastSeen,
        })
    })
    return err
}

type FriendFeedItem struct {
    FriendID     string
    Status       string
    TrackID      *string
    LastActive   int64
    SortPriority int64
}

func (g *GraphRepository) GetSortedFriendsFeed(ctx context.Context, userID string, limit, offset int32) ([]FriendFeedItem, error) {
    l, o := utils.SanitizePagination(limit, offset)

    session := g.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
    defer session.Close(ctx)

    res, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
                query := `
            MATCH (me:User {id: $user_id})-[:FOLLOWS]->(friend:User)
            OPTIONAL MATCH (friend)-[r:LISTENS_TO]->(t:Track)
            
            WITH friend, t, r
            ORDER BY r.last_played DESC
            WITH friend, collect({id: t.id, time: r.last_played})[0] AS last_listen
            
            WITH friend,
                 CASE 
                     WHEN friend.status = 'online' AND friend.current_track_id IS NOT NULL 
                         THEN friend.current_track_id
                     WHEN (friend.status = 'offline' OR friend.status IS NULL) AND last_listen IS NOT NULL
                         THEN last_listen.id
                     ELSE NULL 
                 END AS track_id,
                 CASE 
                     WHEN (friend.status = 'offline' OR friend.status IS NULL) AND last_listen IS NOT NULL
                         THEN last_listen.time
                     ELSE friend.last_seen 
                 END AS last_active,
                 CASE 
                     WHEN friend.status = 'online' AND friend.current_track_id IS NOT NULL THEN 1 
                     WHEN friend.status = 'online' AND friend.current_track_id IS NULL THEN 2     
                     ELSE 3                                                                       
                 END AS sort_priority
            
            ORDER BY sort_priority ASC, friend.id ASC
            SKIP $offset LIMIT $limit
            
            RETURN friend.id AS friend_id, friend.status AS status, track_id, last_active, sort_priority
        `

        result, err := tx.Run(ctx, query, map[string]any{
            "user_id": userID, "offset": o, "limit": l,
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
            lastActive, _ := record.Get("last_active")
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
            
            var lAct int64
            if lastActive != nil {
                lAct = lastActive.(int64)
            }

            feed = append(feed, FriendFeedItem{
                FriendID:     fID.(string),
                Status:       statusStr,
                TrackID:      tID,
                LastActive:   lAct,
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