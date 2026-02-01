package music

import (
    "time"

    "github.com/lib/pq"
)

type Track struct {
    ID          string         `db:"id" json:"id"`
    Title       string         `db:"title" json:"title"`
    Artist      string         `db:"artist" json:"artist"`
    Album       *string        `db:"album" json:"album,omitempty"`
    Duration    int            `db:"duration" json:"duration"`
    Genre       pq.StringArray `db:"genre" json:"genre,omitempty"`
    ImageURL    *string        `db:"image_url" json:"image_url,omitempty"`
    PreviewURL  *string        `db:"preview_url" json:"preview_url,omitempty"`
    VideoID     *string        `db:"video_id" json:"video_id,omitempty"`
    Listeners   int            `db:"listeners" json:"listeners"`
    PlayCount   int            `db:"play_count" json:"play_count"`
    Popularity  int            `db:"popularity" json:"popularity"`
    CreatedAt   time.Time      `db:"created_at" json:"created_at"`
}

type HistoryTrack struct {
    ID       int64     `db:"id" json:"id"`
    UserID   string    `db:"user_id" json:"user_id"`
    TrackID  string    `db:"track_id" json:"track_id"`
    PlayedAt time.Time `db:"played_at" json:"played_at"`
}

type UserTrack struct {
    UserID          string    `db:"user_id" json:"user_id"`
    TrackID         string    `db:"track_id" json:"track_id"`
    InteractionType string    `db:"interaction_type" json:"interaction_type"`
    InteractedAt    time.Time `db:"interacted_at" json:"interacted_at"`
}