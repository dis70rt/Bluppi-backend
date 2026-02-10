package music

import (
    "time"
)

type Track struct {
    ID          string    `db:"track_id"`
    Title       string    `db:"title"`
    Artists     string    `db:"artists"`
    DurationMS  int       `db:"duration_ms"`
    Genres      string    `db:"genres"`
    ImageSmall  *string   `db:"image_small"`
    ImageLarge  *string   `db:"image_large"`
    PreviewURL  *string   `db:"preview_url"`
    VideoID     *string   `db:"video_id"`
    Popularity  int       `db:"popularity"`
    CreatedAt   time.Time `db:"created_at"`

    Listeners int64 `db:"listeners"`
    PlayCount int64 `db:"play_count"`
}

type SearchTrack struct {
    ID         string  `db:"track_id"`
    Title      string  `db:"title"`
    Artists    string  `db:"artists"`
    ImageSmall *string `db:"image_small"`
    PreviewURL *string `db:"preview_url"`
}

type TrackStats struct {
    TrackID   string `db:"track_id"`
    PlayCount int64  `db:"play_count"`
    Listeners int64  `db:"listeners"`
}

type HistoryTrack struct {
    ID       int64     `db:"id"`
    UserID   string    `db:"user_id"`
    TrackID  string    `db:"track_id"`
    PlayedAt time.Time `db:"played_at"`
}

type HistoryEntry struct {
    TrackID    string    `db:"track_id"`
    PlayedAt   time.Time `db:"played_at"`
    Title      string    `db:"title"`
    Artists    string    `db:"artists"`
    ImageSmall *string   `db:"image_small"`
}

type UserTrack struct {
    UserID          string    `db:"user_id"`
    TrackID         string    `db:"track_id"`
    InteractionType string    `db:"interaction_type"`
    InteractedAt    time.Time `db:"interacted_at"`
}