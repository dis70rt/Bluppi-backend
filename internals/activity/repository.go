package activity

import (
    "context"

    "github.com/jmoiron/sqlx"
    "github.com/lib/pq"
)

type Repository struct {
    db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
    return &Repository{db: db}
}

type HydratedUser struct {
    ID         string  `db:"id"`
    Name       string  `db:"name"`
    ProfilePic *string `db:"profile_pic"`
    Username   string  `db:"username"`
}

type HydratedTrack struct {
    TrackID    string  `db:"track_id"`
    Title      string  `db:"title"`
    Artists    string  `db:"artists"`
    ImageSmall *string `db:"image_small"`
    PreviewURL *string `db:"preview_url"`
}

func (r *Repository) GetUsersByIDs(ctx context.Context, ids []string) (map[string]HydratedUser, error) {
    if len(ids) == 0 {
        return make(map[string]HydratedUser), nil
    }

    var users []HydratedUser
    err := r.db.SelectContext(ctx, &users, `SELECT id, name, profile_pic, username FROM users WHERE id = ANY($1)`, pq.Array(ids))
    if err != nil {
        return nil, err
    }

    userMap := make(map[string]HydratedUser)
    for _, u := range users {
        userMap[u.ID] = u
    }
    return userMap, nil
}

func (r *Repository) GetTracksByIDs(ctx context.Context, ids []string) (map[string]HydratedTrack, error) {
    if len(ids) == 0 {
        return make(map[string]HydratedTrack), nil
    }

    var tracks []HydratedTrack
    err := r.db.SelectContext(ctx, &tracks, `SELECT track_id, title, artists, image_small, preview_url FROM tracks WHERE track_id = ANY($1)`, pq.Array(ids))
    if err != nil {
        return nil, err
    }

    trackMap := make(map[string]HydratedTrack)
    for _, t := range tracks {
        trackMap[t.TrackID] = t
    }
    return trackMap, nil
}