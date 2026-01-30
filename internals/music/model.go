package music

import (
    "time"

    "github.com/lib/pq"
)

type Track struct {
    ID          string         `db:"id"`
    Title       string         `db:"title"`
    Artist      string         `db:"artist"`
    Album       *string        `db:"album"`
    Duration    int            `db:"duration"`
    Genre       pq.StringArray `db:"genre"`
    ImageURL    *string        `db:"image_url"`
    PreviewURL  *string        `db:"preview_url"`
    VideoID     *string        `db:"video_id"`
    Listeners   int            `db:"listeners"`
    PlayCount   int            `db:"play_count"`
    Popularity  int            `db:"popularity"`
    CreatedAt   time.Time      `db:"created_at"`
}

type HistoryTrack struct {
    ID       int64     `db:"id"`
    UserID   string    `db:"user_id"`
    TrackID  string    `db:"track_id"`
    PlayedAt time.Time `db:"played_at"`
}

type UserTrack struct {
    UserID          string    `db:"user_id"`
    TrackID         string    `db:"track_id"`
    InteractionType string    `db:"interaction_type"`
    InteractedAt    time.Time `db:"interacted_at"`
}