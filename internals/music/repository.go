package music

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/jmoiron/sqlx"
)

type Querier interface {
    sqlx.QueryerContext
    sqlx.ExecerContext
    GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
    SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
    Rebind(query string) string
    NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

type Repository struct {
    db Querier
}

func NewRepository(db *sqlx.DB) *Repository {
    return &Repository{db: db}
}

func NewRepositoryWithTx(tx *sqlx.Tx) *Repository {
    return &Repository{db: tx}
}

// ----------------- Core Track CRUD -----------------

func (r *Repository) CreateTrack(ctx context.Context, t *Track) error {
    query := `
        INSERT INTO tracks (
            id, title, artist, album, duration, genre,
            image_url, preview_url, video_id,
            listeners, play_count, popularity, created_at
        ) VALUES (
            :id, :title, :artist, :album, :duration, :genre,
            :image_url, :preview_url, :video_id,
            :listeners, :play_count, :popularity, :created_at
        )
        ON CONFLICT (id) DO NOTHING
    `
    _, err := r.db.NamedExecContext(ctx, query, t)
    return err
}

func (r *Repository) GetTrack(ctx context.Context, id string) (*Track, error) {
    var t Track
    err := r.db.GetContext(ctx, &t, `SELECT * FROM tracks WHERE id = $1`, id)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    }
    return &t, err
}

func (r *Repository) UpdateTrack(ctx context.Context, id string, fields map[string]any) error {
    if len(fields) == 0 {
        return errors.New("no fields to update")
    }

    var setClauses []string
    var args []interface{}
    i := 1

    for key, value := range fields {
        setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, i))
        args = append(args, value)
        i++
    }

    args = append(args, id)
    query := fmt.Sprintf("UPDATE tracks SET %s WHERE id = $%d", strings.Join(setClauses, ", "), i)

    res, err := r.db.ExecContext(ctx, query, args...)
    if err != nil {
        return err
    }

    rows, err := res.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return sql.ErrNoRows
    }
    return nil
}

func (r *Repository) DeleteTrack(ctx context.Context, id string) error {
    res, err := r.db.ExecContext(ctx, `DELETE FROM tracks WHERE id = $1`, id)
    if err != nil {
        return err
    }
    if rows, _ := res.RowsAffected(); rows == 0 {
        return sql.ErrNoRows
    }
    return nil
}

// ----------------- Search & Discovery -----------------

func (r *Repository) SearchTracks(
    ctx context.Context,
    query string,
    limit, offset int,
) ([]Track, int, error) {
    search := "%" + query + "%"
    tracks := []Track{}

    err := r.db.SelectContext(
        ctx,
        &tracks,
        `
        SELECT * FROM tracks 
        WHERE title ILIKE $1 OR artist ILIKE $1 OR album ILIKE $1
        ORDER BY popularity DESC
        LIMIT $2 OFFSET $3
        `,
        search, limit, offset,
    )
    if err != nil {
        return nil, 0, err
    }

    var total int
    err = r.db.GetContext(
        ctx,
        &total,
        `SELECT COUNT(*) FROM tracks WHERE title ILIKE $1 OR artist ILIKE $1 OR album ILIKE $1`,
        search,
    )

    return tracks, total, err
}

func (r *Repository) GetPopularTracks(ctx context.Context, limit int) ([]Track, error) {
    tracks := []Track{}
    err := r.db.SelectContext(
        ctx,
        &tracks,
        `SELECT * FROM tracks ORDER BY popularity DESC LIMIT $1`,
        limit,
    )
    return tracks, err
}

func (r *Repository) GetTracksByGenre(
    ctx context.Context,
    genre string,
    limit, offset int,
) ([]Track, int, error) {
    tracks := []Track{}

    err := r.db.SelectContext(
        ctx,
        &tracks,
        `
        SELECT * FROM tracks
        WHERE $1 = ANY(genre)
        ORDER BY popularity DESC
        LIMIT $2 OFFSET $3
        `,
        genre, limit, offset,
    )
    if err != nil {
        return nil, 0, err
    }

    var total int
    err = r.db.GetContext(
        ctx,
        &total,
        `SELECT COUNT(*) FROM tracks WHERE $1 = ANY(genre)`,
        genre,
    )

    return tracks, total, err
}

// ----------------- User Interactions (Likes) -----------------

type LikedTrackEntry struct {
    TrackID    string    `db:"track_id"`
    LikedAt    time.Time `db:"interacted_at"`
    Title      string    `db:"title"`
    Artist     string    `db:"artist"`
    Album      *string   `db:"album"`
    ImageURL   *string   `db:"image_url"`
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
        SELECT ut.track_id, ut.interacted_at, t.title, t.artist, t.album, t.image_url
        FROM user_track ut
        JOIN tracks t ON ut.track_id = t.id
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

type HistoryEntry struct {
    TrackID  string    `db:"track_id"`
    PlayedAt time.Time `db:"played_at"`
    Title    string    `db:"title"`
    Artist   string    `db:"artist"`
    Album    *string   `db:"album"`
    ImageURL *string   `db:"image_url"`
}

func (r *Repository) AddTrackToHistory(ctx context.Context, userID, trackID string) error {
    _, err := r.db.ExecContext(
        ctx,
        `INSERT INTO history_tracks (user_id, track_id, played_at) VALUES ($1, $2, NOW())`,
        userID, trackID,
    )
    if err != nil {
        return err
    }

    _, err = r.db.ExecContext(
        ctx,
        `
        INSERT INTO user_track (user_id, track_id, interaction_type, interacted_at)
        VALUES ($1, $2, 'last_played', NOW())
        ON CONFLICT (user_id, track_id, interaction_type)
        DO UPDATE SET interacted_at = NOW()
        `,
        userID, trackID,
    )

    return err
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
        SELECT h.track_id, h.played_at, t.title, t.artist, t.album, t.image_url
        FROM history_tracks h
        JOIN tracks t ON h.track_id = t.id
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