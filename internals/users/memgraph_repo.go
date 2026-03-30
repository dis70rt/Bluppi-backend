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

func (g *GraphRepository) GetSuggestedUsers(ctx context.Context, userID string, limit int, cursor string) ([]string, string, error) {
	session := g.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)
  
	cursorTier, cursorWeight, cursorID := utils.DecodeCursor(cursor)
	hasCursor := cursor != ""
	fallbackLimit := limit * 2
  
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
            MATCH (u:User {id: $user_id})
 
            // 1. Mutuals
            CALL {
                WITH u
                MATCH (u)-[:FOLLOWS]->(mutual:User)-[:FOLLOWS]->(s1:User)
                WHERE s1 <> u AND NOT (u)-[:FOLLOWS]->(s1)
                WITH s1, count(mutual) AS mutual_count
                ORDER BY mutual_count DESC
                LIMIT 100
                WITH collect(s1) AS nodes
                RETURN nodes
            }
            WITH u, nodes AS mutual_nodes
 
            // 2. Shared Interests
            CALL {
                WITH u
                MATCH (u)-[:LIKES|LISTENS_TO]->(t:Track)<-[:LIKES|LISTENS_TO]-(s2:User)
                WHERE s2 <> u AND NOT (u)-[:FOLLOWS]->(s2)
                WITH s2, count(t) AS shared_track_count
                ORDER BY shared_track_count DESC
                LIMIT 100
                WITH collect(s2) AS nodes
                RETURN nodes
            }
            WITH u, mutual_nodes, nodes AS shared_nodes
 
            // 3. Followers I don't follow back
            CALL {
                WITH u
                MATCH (s3:User)-[:FOLLOWS]->(u)
                WHERE NOT (u)-[:FOLLOWS]->(s3)
                WITH s3
                ORDER BY s3.id DESC
                LIMIT 100
                WITH collect(s3) AS nodes
                RETURN nodes
            }
            WITH u, mutual_nodes, shared_nodes, nodes AS follower_nodes
 
            // 4. Tag and Combine
            WITH u,
                 [m IN mutual_nodes   WHERE m IS NOT NULL | {node: m, type: 'mutual'}]   AS m_list,
                 [s IN shared_nodes   WHERE s IS NOT NULL | {node: s, type: 'shared'}]   AS s_list,
                 [f IN follower_nodes WHERE f IS NOT NULL | {node: f, type: 'follower'}] AS f_list
            WITH u, m_list + s_list + f_list AS all_connections
 
            // 5. UNWIND and GROUP by node.id for stable deduplication
            UNWIND (CASE WHEN size(all_connections) = 0 THEN [{node: null, type: null}] ELSE all_connections END) AS conn
            WITH u,
                 conn.node          AS suggested,
                 conn.node.id       AS suggested_id_key,
                 collect(conn.type) AS types,
                 count(conn)        AS weight
 
            // 6. Assign Tiers & Collect Safely
            WITH u, suggested, weight,
                 CASE
                    WHEN suggested IS NULL                           THEN 0
                    WHEN 'mutual' IN types AND 'shared' IN types     THEN 5
                    WHEN 'mutual' IN types                           THEN 4
                    WHEN 'shared' IN types                           THEN 3
                    WHEN 'follower' IN types                         THEN 2
                    ELSE 0
                 END AS tier
 
            WITH u,
                 collect(CASE WHEN suggested IS NOT NULL
                              THEN {node: suggested, tier: tier, weight: weight}
                              ELSE null
                         END) AS raw_personalized
            WITH u, [p IN raw_personalized WHERE p IS NOT NULL] AS personalized
 
            // 7. Global Fallback
            OPTIONAL MATCH (fallback:User)
            WHERE fallback <> u AND NOT (u)-[:FOLLOWS]->(fallback)
 
            WITH u, personalized, fallback
            ORDER BY fallback.id DESC
            LIMIT $fallback_limit
 
            WITH personalized,
                 collect(CASE WHEN fallback IS NOT NULL
                              THEN {node: fallback, tier: 1, weight: 0}
                              ELSE null
                         END) AS raw_general
            WITH personalized, [g IN raw_general WHERE g IS NOT NULL] AS general
 
            UNWIND (CASE WHEN size(personalized + general) = 0 THEN [null] ELSE personalized + general END) AS candidate
            WITH candidate
            WHERE candidate IS NOT NULL
 
            WITH candidate.node        AS target,
                 candidate.node.id     AS target_id_key,
                 max(candidate.tier)   AS final_tier,
                 max(candidate.weight) AS final_weight
            WHERE target IS NOT NULL
              AND (
                    NOT $has_cursor
                    OR final_tier < $cursor_tier
                    OR (final_tier = $cursor_tier AND final_weight < $cursor_weight)
                    OR (final_tier = $cursor_tier AND final_weight = $cursor_weight AND target.id < $cursor_id)
                  )
 
            RETURN target.id AS suggested_id, final_tier, final_weight
            ORDER BY final_tier DESC, final_weight DESC, target.id DESC
            LIMIT $limit
        `
 
		params := map[string]any{
			"user_id":        userID,
			"limit":          limit,
			"fallback_limit": fallbackLimit,
			"has_cursor":     hasCursor,
			"cursor_tier":    cursorTier,
			"cursor_weight":  cursorWeight,
			"cursor_id":      cursorID,
		}
 
		records, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
 
		return records.Collect(ctx)
	})
 
	if err != nil {
		return nil, "", err
	}
 
	records := result.([]*neo4j.Record)
 
	var suggestedUsers []string
	var nextCursor string
 
	for i, record := range records {
		suggestedID, _ := record.Get("suggested_id")
		tier, _ := record.Get("final_tier")
		weight, _ := record.Get("final_weight")
  
		suggestedUsers = append(suggestedUsers, suggestedID.(string))
 
		if i == len(records)-1 {
			nextCursor = utils.EncodeCursor(int(tier.(int64)), int(weight.(int64)), suggestedID.(string))
		}
	}
 
	if len(suggestedUsers) < limit {
		nextCursor = ""
	}
 
	return suggestedUsers, nextCursor, nil
}