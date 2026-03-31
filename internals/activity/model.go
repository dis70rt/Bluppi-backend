package activity

import "errors"

var (
    ErrInvalidInput = errors.New("invalid input")
)

type Activity struct {
    FriendID        string
    FriendName      string
    FriendAvatarURL string
    Status          string
    FriendUsername  string
    TrackID         *string
    TrackTitle      *string
    TrackArtist     *string
    TrackCoverURL   *string
    TrackPreviewURL *string
    LastSeen        int64
}