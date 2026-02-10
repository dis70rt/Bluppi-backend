package music

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Querier interface {
    sqlx.QueryerContext
    sqlx.ExecerContext
    GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
    SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
    Rebind(query string) string
    NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

type SolrSearcher interface {
    Search(ctx context.Context, query string, limit int, cursor string) ([]string, string, error)
}

type Repository struct {
    db Querier
    solr SolrSearcher
}

func NewRepository(db *sqlx.DB, solr SolrSearcher) *Repository {
    return &Repository{db: db, solr: solr}
}

func NewRepositoryWithTx(tx *sqlx.Tx, solr SolrSearcher) *Repository {
    return &Repository{db: tx, solr: solr}
}

func baseSelectQuery() string {
    return `
        SELECT 
            t.track_id, t.title, t.artists, t.genres, t.duration_ms, 
            t.image_small, t.image_large, t.preview_url, t.video_id, 
            t.popularity, t.created_at,
            COALESCE(s.listeners, 0) as listeners,
            COALESCE(s.play_count, 0) as play_count
        FROM tracks t
        LEFT JOIN track_stats s ON t.track_id = s.track_id
    `
}

// ----------------- Core Track Reading -----------------

func (r *Repository) GetTrack(ctx context.Context, id string) (*Track, error) {
    var t Track
    query := baseSelectQuery() + ` WHERE t.track_id = $1`
    err := r.db.GetContext(ctx, &t, query, id)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    }
    return &t, err
}

// ----------------- Search & Discovery -----------------

func (r *Repository) SearchTracks(
    ctx context.Context,
    query string,
    limit int,
    cursor string,
) ([]SearchTrack, string, error) {

    solrIDs, nextCursor, err := r.solr.Search(ctx, query, limit, cursor)
    if err != nil {
        return nil, "", err
    }

    if len(solrIDs) == 0 {
        return []SearchTrack{}, nextCursor, nil
    }

    tracks := []SearchTrack{}
    err = r.db.SelectContext(
        ctx,
        &tracks,
        `
        SELECT track_id, title, artists, image_small, preview_url
        FROM tracks
        WHERE track_id = ANY($1)
        `,
        pq.Array(solrIDs),
    )
    if err != nil {
        return nil, "", err
    }

    reorderTracks(tracks, solrIDs)
    return tracks, nextCursor, err
}

func reorderTracks(tracks []SearchTrack, solrIDs []string) {
    order := make(map[string]int, len(solrIDs))
    for i, id := range solrIDs {
        order[id] = i
    }

    sort.SliceStable(tracks, func(i, j int) bool {
        return order[tracks[i].ID] < order[tracks[j].ID]
    })
}

func (r *Repository) GetPopularTracks(ctx context.Context, limit int) ([]Track, error) {
    tracks := []Track{}
    query := baseSelectQuery() + ` ORDER BY t.popularity DESC LIMIT $1`
    
    err := r.db.SelectContext(ctx, &tracks, query, limit)
    return tracks, err
}

func (r *Repository) GetTracksByGenre(
    ctx context.Context,
    genre string,
    limit, offset int,
) ([]Track, int, error) {
    tracks := []Track{}

    searchPattern := "%" + genre + "%"

    query := baseSelectQuery() + `
        WHERE t.genres ILIKE $1
        ORDER BY t.popularity DESC
        LIMIT $2 OFFSET $3
    `

    err := r.db.SelectContext(ctx, &tracks, query, searchPattern, limit, offset)
    if err != nil {
        return nil, 0, err
    }

    var total int
    err = r.db.GetContext(
        ctx,
        &total,
        `SELECT COUNT(*) FROM tracks WHERE genres ILIKE $1`,
        searchPattern,
    )

    return tracks, total, err
}

// ----------------- User Interactions (Likes) -----------------

type LikedTrackEntry struct {
    TrackID    string    `db:"track_id"`
    LikedAt    time.Time `db:"interacted_at"`
    Title      string    `db:"title"`
    Artists    string    `db:"artists"`
    ImageSmall *string   `db:"image_small"`
}

func (r *Repository) LikeTrack(ctx context.Context, userID, trackID string) error {
    _, err := r.db.ExecContext(
        ctx,
        `
        INSERT INTO user_track (user_id, track_id, interaction_type)
        VALUES ($1, $2, 'liked')
        ON CONFLICT (user_id, track_id, interaction_type) DO NOTHING
        `,
        userID, trackID,
    )
    return err
}

func (r *Repository) UnlikeTrack(ctx context.Context, userID, trackID string) error {
    res, err := r.db.ExecContext(
        ctx,
        `DELETE FROM user_track WHERE user_id = $1 AND track_id = $2 AND interaction_type = 'liked'`,
        userID, trackID,
    )
    if err != nil {
        return err
    }
    if rows, _ := res.RowsAffected(); rows == 0 {
        return sql.ErrNoRows
    }
    return nil
}

func (r *Repository) IsTrackLiked(ctx context.Context, userID, trackID string) (bool, error) {
    var exists bool
    err := r.db.GetContext(
        ctx,
        &exists,
        `SELECT EXISTS (
            SELECT 1 FROM user_track 
            WHERE user_id = $1 AND track_id = $2 AND interaction_type = 'liked'
        )`,
        userID, trackID,
    )
    return exists, err
}

func (r *Repository) GetLikedTracks(
    ctx context.Context,
    userID string,
    limit, offset int,
) ([]LikedTrackEntry, int, error) {
    results := []LikedTrackEntry{}

    err := r.db.SelectContext(
        ctx,
        &results,
        `
        SELECT ut.track_id, ut.interacted_at, t.title, t.artists, t.image_small
        FROM user_track ut
        JOIN tracks t ON ut.track_id = t.track_id
        WHERE ut.user_id = $1 AND ut.interaction_type = 'liked'
        ORDER BY ut.interacted_at DESC
        LIMIT $2 OFFSET $3
        `,
        userID, limit, offset,
    )
    if err != nil {
        return nil, 0, err
    }

    var total int
    err = r.db.GetContext(
        ctx,
        &total,
        `SELECT COUNT(*) FROM user_track WHERE user_id = $1 AND interaction_type = 'liked'`,
        userID,
    )

    return results, total, err
}

// ----------------- History -----------------

func (r *Repository) AddTrackToHistory(ctx context.Context, userID, trackID string) error {
    _, err := r.db.ExecContext(
        ctx,
        `INSERT INTO history_tracks (user_id, track_id, played_at) VALUES ($1, $2, NOW())`,
        userID, trackID,
    )
    if err != nil {
        return fmt.Errorf("failed to add history log: %w", err)
    }

    return nil
}

func (r *Repository) GetTrackHistory(
    ctx context.Context,
    userID string,
    limit, offset int,
) ([]HistoryEntry, int, error) {
    history := []HistoryEntry{}

    err := r.db.SelectContext(
        ctx,
        &history,
        `
        SELECT h.track_id, h.played_at, t.title, t.artists, t.image_small
        FROM history_tracks h
        JOIN tracks t ON h.track_id = t.track_id
        WHERE h.user_id = $1
        ORDER BY h.played_at DESC
        LIMIT $2 OFFSET $3
        `,
        userID, limit, offset,
    )
    if err != nil {
        return nil, 0, err
    }

    var total int
    err = r.db.GetContext(
        ctx,
        &total,
        `SELECT COUNT(*) FROM history_tracks WHERE user_id = $1`,
        userID,
    )

    return history, total, err
}

func (r *Repository) ClearTrackHistory(ctx context.Context, userID string) error {
    _, err := r.db.ExecContext(ctx, `DELETE FROM history_tracks WHERE user_id = $1`, userID)
    return err
}