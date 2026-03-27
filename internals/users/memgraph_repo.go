package users

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type GraphRepository struct {
	driver neo4j.DriverWithContext
}

func NewGraphRepository(driver neo4j.DriverWithContext) *GraphRepository {
	return &GraphRepository{driver: driver}
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

func (g *GraphRepository) GetSuggestedUsers(ctx context.Context, userID string, limit int) ([]string, error) {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
            MATCH (u:User {id: $user_id})
            
            OPTIONAL MATCH (u)-[:FOLLOWS]->(:User)-[:FOLLOWS]->(s1:User)
            WHERE s1 <> u AND NOT (u)-[:FOLLOWS]->(s1)
            WITH u, collect(s1) AS social_suggests
            
            OPTIONAL MATCH (u)-[:LIKES|LISTENS_TO]->(:Track)<-[:LIKES|LISTENS_TO]-(s2:User)
            WHERE s2 <> u AND NOT (u)-[:FOLLOWS]->(s2)
            WITH social_suggests + collect(s2) AS all_suggests
            
            UNWIND all_suggests AS suggested
            WITH suggested, count(suggested) AS score
            WHERE suggested IS NOT NULL
            RETURN suggested.id AS suggested_id
            ORDER BY score DESC
            LIMIT $limit
        `
		records, err := tx.Run(ctx, query, map[string]any{
			"user_id": userID,
			"limit":   limit,
		})
		if err != nil {
			return nil, err
		}

		var suggestedUsers []string
		for records.Next(ctx) {
			record := records.Record()
			suggestedID, ok := record.Get("suggested_id")
			if ok && suggestedID != nil {
				suggestedUsers = append(suggestedUsers, suggestedID.(string))
			}
		}
		return suggestedUsers, records.Err()
	})

	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}
