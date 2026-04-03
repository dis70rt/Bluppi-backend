package notifications

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type DeviceToken struct {
    ID         string    `db:"id"`
    UserID     string    `db:"user_id"`
    FCMToken   string    `db:"fcm_token"`
    DeviceType string    `db:"device_type"`
    IsActive   bool      `db:"is_active"`
    CreatedAt  time.Time `db:"created_at"`
    LastUsedAt time.Time `db:"last_used_at"`
}

type NotificationPreferences struct {
    UserID                    string    `db:"user_id"`
    PushNotificationsEnabled  bool      `db:"push_notifications_enabled"`
    EmailNotificationsEnabled bool      `db:"email_notifications_enabled"`
    PartyInvitesEnabled       bool      `db:"party_invites_enabled"`
    NewFollowersEnabled       bool      `db:"new_followers_enabled"`
    FollowRequestEnabled      bool      `db:"follow_request_enabled"`
    FollowerListeningEnabled  bool      `db:"follower_listening_enabled"`
    UpdatedAt                 time.Time `db:"updated_at"`
}

type NotificationType string

const (
    NotificationTypePartyInvite         NotificationType = "PARTY_INVITE"
    NotificationTypeNewFollower         NotificationType = "NEW_FOLLOWER"
    NotificationTypeFollowRequest       NotificationType = "FOLLOW_REQUEST"
    NotificationTypeFollowerListening   NotificationType = "FOLLOWER_LISTENING"
)

type ActionData map[string]interface{}

func (ad ActionData) Value() (driver.Value, error) {
    return json.Marshal(ad)
}

func (ad *ActionData) Scan(value interface{}) error {
    bytes, ok := value.([]byte)
    if !ok {
        return errors.New("type assertion failed")
    }
    return json.Unmarshal(bytes, &ad)
}

type NotificationHistory struct {
    ID        string           `db:"id"`
    UserID    string           `db:"user_id"`
    Type      NotificationType `db:"type"`
    Title     string           `db:"title"`
    Body      string           `db:"body"`
    ActionData ActionData      `db:"action_data"`
    IsRead    bool             `db:"is_read"`
    CreatedAt time.Time        `db:"created_at"`
}