package notifications
import (
    "context"
    "database/sql"
    "errors"
    "time"
)

var (
    ErrInvalidInput          = errors.New("invalid input")
    ErrNotificationNotFound  = errors.New("notification not found")
    ErrDeviceNotFound        = errors.New("device not found")
    ErrPreferencesNotFound   = errors.New("notification preferences not found")
    ErrInvalidDeviceType     = errors.New("invalid device type")
    ErrNoNotificationIDs     = errors.New("no notification IDs provided")
)

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}


func (s *Service) RegisterDevice(ctx context.Context, userID, fcmToken, deviceType string) error {
    if userID == "" || fcmToken == "" || deviceType == "" {
        return ErrInvalidInput
    }
    if !isValidDeviceType(deviceType) {
        return ErrInvalidDeviceType
    }

    now := time.Now().UTC()
    dt := &DeviceToken{
        UserID:     userID,
        FCMToken:   fcmToken,
        DeviceType: deviceType,
        IsActive:   true,
        CreatedAt:  now,
        LastUsedAt: now,
    }

    return s.repo.RegisterDevice(ctx, dt)
}

func (s *Service) UnregisterDevice(ctx context.Context, fcmToken string) error {
    if fcmToken == "" {
        return ErrInvalidInput
    }

    err := s.repo.UnregisterDevice(ctx, fcmToken)
    if errors.Is(err, sql.ErrNoRows) {
        return ErrDeviceNotFound
    }
    return err
}

func (s *Service) GetDeviceTokensByUserID(ctx context.Context, userID string) ([]DeviceToken, error) {
    if userID == "" {
        return nil, ErrInvalidInput
    }
    return s.repo.GetDeviceTokensByUserID(ctx, userID)
}

func (s *Service) GetActiveDeviceTokensByUserID(ctx context.Context, userID string) ([]DeviceToken, error) {
    if userID == "" {
        return nil, ErrInvalidInput
    }
    return s.repo.GetActiveDeviceTokensByUserID(ctx, userID)
}

func (s *Service) DeactivateDevice(ctx context.Context, fcmToken string) error {
    if fcmToken == "" {
        return ErrInvalidInput
    }

    err := s.repo.DeactivateDevice(ctx, fcmToken)
    if errors.Is(err, sql.ErrNoRows) {
        return ErrDeviceNotFound
    }
    return err
}

func (s *Service) UpdateDeviceLastUsed(ctx context.Context, fcmToken string) error {
    if fcmToken == "" {
        return ErrInvalidInput
    }

    err := s.repo.UpdateDeviceLastUsed(ctx, fcmToken)
    if errors.Is(err, sql.ErrNoRows) {
        return ErrDeviceNotFound
    }
    return err
}

// ----- Preferences Operations -----

func (s *Service) GetPreferences(ctx context.Context, userID string) (*NotificationPreferences, error) {
    if userID == "" {
        return nil, ErrInvalidInput
    }

    prefs, err := s.repo.GetPreferences(ctx, userID)
    if err != nil {
        return nil, err
    }
    if prefs == nil {
        return nil, ErrPreferencesNotFound
    }
    return prefs, nil
}

func (s *Service) UpdatePreferencesFields(ctx context.Context, userID string, fields map[string]any) error {
    if userID == "" || len(fields) == 0 {
        return ErrInvalidInput
    }

    err := s.repo.UpdatePreferences(ctx, userID, fields)
    if errors.Is(err, sql.ErrNoRows) {
        return ErrPreferencesNotFound
    }
    return err
}

func (s *Service) UpdatePreferences(ctx context.Context, userID string, prefs NotificationPreferences) error {
    if userID == "" {
        return ErrInvalidInput
    }

    fields := map[string]any{
        "push_notifications_enabled":   prefs.PushNotificationsEnabled,
        "email_notifications_enabled":  prefs.EmailNotificationsEnabled,
        "party_invites_enabled":        prefs.PartyInvitesEnabled,
        "new_followers_enabled":        prefs.NewFollowersEnabled,
        "follow_request_enabled":       prefs.FollowRequestEnabled,
        "follower_listening_enabled":   prefs.FollowerListeningEnabled,
    }

    err := s.repo.UpdatePreferences(ctx, userID, fields)
    if errors.Is(err, sql.ErrNoRows) {
        return ErrPreferencesNotFound
    }
    return err
}

// ----- Notification History Operations -----

func (s *Service) CreateNotification(
    ctx context.Context,
    userID string,
    typ NotificationType,
    title string,
    body string,
    actionData ActionData,
) error {
    if userID == "" || typ == "" || title == "" || body == "" {
        return ErrInvalidInput
    }

    nh := &NotificationHistory{
        UserID:     userID,
        Type:       typ,
        Title:      title,
        Body:       body,
        ActionData: actionData,
        IsRead:     false,
        CreatedAt:  time.Now().UTC(),
    }

    return s.repo.CreateNotification(ctx, nh)
}

func (s *Service) GetNotificationByID(ctx context.Context, notificationID string) (*NotificationHistory, error) {
    if notificationID == "" {
        return nil, ErrInvalidInput
    }

    nh, err := s.repo.GetNotificationByID(ctx, notificationID)
    if err != nil {
        return nil, err
    }
    if nh == nil {
        return nil, ErrNotificationNotFound
    }
    return nh, nil
}

func (s *Service) GetHistory(ctx context.Context, userID string, limit, offset int) ([]NotificationHistory, int, error) {
    if userID == "" {
        return nil, 0, ErrInvalidInput
    }
    limit, offset = normalizePagination(limit, offset)

    list, err := s.repo.GetNotificationHistory(ctx, userID, limit, offset)
    if err != nil {
        return nil, 0, err
    }

    unread, err := s.repo.GetUnreadCount(ctx, userID)
    if err != nil {
        return nil, 0, err
    }

    return list, unread, nil
}

func (s *Service) GetUnreadNotifications(ctx context.Context, userID string, limit, offset int) ([]NotificationHistory, error) {
    if userID == "" {
        return nil, ErrInvalidInput
    }
    limit, offset = normalizePagination(limit, offset)

    return s.repo.GetUnreadNotifications(ctx, userID, limit, offset)
}

func (s *Service) GetUnreadCount(ctx context.Context, userID string) (int, error) {
    if userID == "" {
        return 0, ErrInvalidInput
    }
    return s.repo.GetUnreadCount(ctx, userID)
}

func (s *Service) MarkAsRead(ctx context.Context, notificationIDs []string) error {
    if len(notificationIDs) == 0 {
        return ErrNoNotificationIDs
    }
    // NOTE: If you want ownership validation by user, add repo method with:
    // UPDATE ... WHERE user_id = $X AND id IN (...)
    return s.repo.MarkAsRead(ctx, notificationIDs)
}

func (s *Service) MarkAllAsRead(ctx context.Context, userID string) error {
    if userID == "" {
        return ErrInvalidInput
    }
    return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *Service) DeleteNotification(ctx context.Context, notificationID string) error {
    if notificationID == "" {
        return ErrInvalidInput
    }

    err := s.repo.DeleteNotification(ctx, notificationID)
    if errors.Is(err, sql.ErrNoRows) {
        return ErrNotificationNotFound
    }
    return err
}

func (s *Service) ClearHistory(ctx context.Context, userID string) error {
    if userID == "" {
        return ErrInvalidInput
    }
    return s.repo.ClearHistory(ctx, userID)
}

func isValidDeviceType(deviceType string) bool {
    switch deviceType {
    case "android", "ios", "web":
        return true
    default:
        return false
    }
}

func normalizePagination(limit, offset int) (int, int) {
    if limit <= 0 || limit > 100 {
        limit = 20
    }
    if offset < 0 {
        offset = 0
    }
    return limit, offset
}