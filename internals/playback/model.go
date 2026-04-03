package playback

import (
    "sync"
    pb "github.com/dis70rt/bluppi-backend/internals/gen/playback"
)

type RoomState struct {
    ID          string                    `json:"id"`
    HostUserID  string                    `json:"host_user_id"`
    mu          sync.RWMutex
    Clients     map[string]*ClientSession `json:"clients"`
    State       PlaybackState             `json:"playback_state"`
    BufferReady map[string]bool           `json:"buffer_ready"`
}

type PlaybackState struct {
    Track                  TrackInfo `json:"track"` 
    IsPlaying              bool      `json:"is_playing"`
    PositionMs             int64     `json:"position_ms"`
    LastUpdateServerUs     int64     `json:"last_update_server_us"`
    ScheduledStartServerUs int64     `json:"scheduled_start_server_us"`
    Version                uint64    `json:"version"`
}

type TrackInfo struct {
    ID         string `json:"id"`
    Title      string `json:"title"`
    Artist     string `json:"artist"`
    AudioURL   string `json:"audio_url"`
    ArtworkURL string `json:"artwork_url"`
    DurationMs int64  `json:"duration_ms"`
}

type ClientSession struct {
    UserID  string `json:"user_id"`
    Send    chan *pb.ServerEvent
}