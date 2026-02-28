package party

import "time"

type RoomStatus string
type RoomVisibility string
type RoomMemberRole string

const (
	RoomStatusActive RoomStatus = "ACTIVE"
	RoomStatusEnded  RoomStatus = "ENDED"

	RoomVisibilityPublic  RoomVisibility = "PUBLIC"
	RoomVisibilityPrivate RoomVisibility = "PRIVATE"

	RoomRoleHost     RoomMemberRole = "HOST"
	RoomRoleListener RoomMemberRole = "LISTENER"
)

type Room struct {
	ID         string         `db:"id"`
	Name       string         `db:"name"`
	Code       string         `db:"code"`
	Status     RoomStatus     `db:"status"`
	Visibility RoomVisibility `db:"visibility"`
	InviteOnly bool           `db:"invite_only"`
	HostUserID string         `db:"host_user_id"`
	CreatedAt  time.Time      `db:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at"`
}

type RoomMember struct {
	RoomID   string          `db:"room_id"`
	UserID   string          `db:"user_id"`
	Role     RoomMemberRole  `db:"role"`
	JoinedAt time.Time       `db:"joined_at"`
	LeftAt   *time.Time      `db:"left_at"`
}

type RoomSummary struct {
	RoomID        string         `db:"room_id"`
	RoomName      string         `db:"room_name"`
	HostUserID    string         `db:"host_user_id"`
	ListenerCount int32          `db:"listener_count"`
	IsLive        bool           `db:"is_live"`
	Visibility    RoomVisibility `db:"visibility"`
}