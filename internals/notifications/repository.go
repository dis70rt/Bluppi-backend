package notifications

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "strings"

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

func (r *Repository) RegisterDevice(ctx context.Context, dt *DeviceToken) error {
    query := `
        INSERT INTO device_tokens (user_id, fcm_token, device_type, is_active, created_at, last_used_at)
        VALUES (:user_id, :fcm_token, :device_type, :is_active, :created_at, :last_used_at)
        ON CONFLICT (fcm_token) DO UPDATE SET
            is_active = EXCLUDED.is_active,
            last_used_at = EXCLUDED.last_used_at
    `
    _, err := r.db.NamedExecContext(ctx, query, dt)
    return err
}

func (r *Repository) UnregisterDevice(ctx context.Context, fcmToken string) error {
    res, err := r.db.ExecContext(
        ctx,
        `DELETE FROM device_tokens WHERE fcm_token = $1`,
        fcmToken,
    )
    if err != nil {
        return err
    }
    if rows, _ := res.RowsAffected(); rows == 0 {
        return sql.ErrNoRows
    }
    return nil
}

func (r *Repository) GetDeviceTokensByUserID(ctx context.Context, userID string) ([]DeviceToken, error) {
    tokens := []DeviceToken{}

    err := r.db.SelectContext(
        ctx,
        &tokens,
        `SELECT * FROM device_tokens WHERE user_id = $1 ORDER BY created_at DESC`,
        userID,
    )

    return tokens, err
}

func (r *Repository) GetActiveDeviceTokensByUserID(ctx context.Context, userID string) ([]DeviceToken, error) {
    tokens := []DeviceToken{}

    err := r.db.SelectContext(
        ctx,
        &tokens,
        `SELECT * FROM device_tokens WHERE user_id = $1 AND is_active = TRUE ORDER BY last_used_at DESC`,
        userID,
    )

    return tokens, err
}

func (r *Repository) GetDeviceTokenByFCMToken(ctx context.Context, fcmToken string) (*DeviceToken, error) {
    var dt DeviceToken

    err := r.db.GetContext(
        ctx,
        &dt,
        `SELECT * FROM device_tokens WHERE fcm_token = $1`,
        fcmToken,
    )

    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    }
    return &dt, err
}

func (r *Repository) UpdateDeviceLastUsed(ctx context.Context, fcmToken string) error {
    res, err := r.db.ExecContext(
        ctx,
        `UPDATE device_tokens SET last_used_at = NOW() WHERE fcm_token = $1`,
        fcmToken,
    )
    if err != nil {
        return err
    }
    if rows, _ := res.RowsAffected(); rows == 0 {
        return sql.ErrNoRows
    }
    return nil
}

func (r *Repository) DeactivateDevice(ctx context.Context, fcmToken string) error {
    res, err := r.db.ExecContext(
        ctx,
        `UPDATE device_tokens SET is_active = FALSE WHERE fcm_token = $1`,
        fcmToken,
    )
    if err != nil {
        return err
    }
    if rows, _ := res.RowsAffected(); rows == 0 {
        return sql.ErrNoRows
    }
    return nil
}

func (r *Repository) GetPreferences(ctx context.Context, userID string) (*NotificationPreferences, error) {
    var prefs NotificationPreferences

    err := r.db.GetContext(
        ctx,
        &prefs,
        `SELECT * FROM notification_preferences WHERE user_id = $1`,
        userID,
    )

    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    }
    return &prefs, err
}

func (r *Repository) UpdatePreferences(ctx context.Context, userID string, fields map[string]any) error {
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

    args = append(args, userID)
    query := fmt.Sprintf("UPDATE notification_preferences SET %s WHERE user_id = $%d", strings.Join(setClauses, ", "), i)

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

func (r *Repository) CreateNotification(ctx context.Context, nh *NotificationHistory) error {
    query := `
        INSERT INTO notification_history (user_id, type, title, body, action_data, is_read, created_at)
        VALUES (:user_id, :type, :title, :body, :action_data, :is_read, :created_at)
    `
    _, err := r.db.NamedExecContext(ctx, query, nh)
    return err
}

func (r *Repository) GetNotificationByID(ctx context.Context, notificationID string) (*NotificationHistory, error) {
    var nh NotificationHistory

    err := r.db.GetContext(
        ctx,
        &nh,
        `SELECT * FROM notification_history WHERE id = $1`,
        notificationID,
    )

    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    }
    return &nh, err
}

func (r *Repository) GetNotificationHistory(ctx context.Context, userID string, limit, offset int) ([]NotificationHistory, error) {
    notifications := []NotificationHistory{}

    err := r.db.SelectContext(
        ctx,
        &notifications,
        `
        SELECT * FROM notification_history
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
        `,
        userID, limit, offset,
    )

    return notifications, err
}

func (r *Repository) GetUnreadNotifications(ctx context.Context, userID string, limit, offset int) ([]NotificationHistory, error) {
    notifications := []NotificationHistory{}

    err := r.db.SelectContext(
        ctx,
        &notifications,
        `
        SELECT * FROM notification_history
        WHERE user_id = $1 AND is_read = FALSE
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
        `,
        userID, limit, offset,
    )

    return notifications, err
}

func (r *Repository) GetUnreadCount(ctx context.Context, userID string) (int, error) {
    var count int

    err := r.db.GetContext(
        ctx,
        &count,
        `SELECT COUNT(*) FROM notification_history WHERE user_id = $1 AND is_read = FALSE`,
        userID,
    )

    return count, err
}

func (r *Repository) MarkAsRead(ctx context.Context, notificationIDs []string) error {
    if len(notificationIDs) == 0 {
        return errors.New("no notification IDs provided")
    }

    query, args, err := sqlx.In(
        `UPDATE notification_history SET is_read = TRUE WHERE id IN (?)`,
        notificationIDs,
    )
    if err != nil {
        return err
    }

    query = r.db.Rebind(query)
    _, err = r.db.ExecContext(ctx, query, args...)
    return err
}

func (r *Repository) MarkAllAsRead(ctx context.Context, userID string) error {
    res, err := r.db.ExecContext(
        ctx,
        `UPDATE notification_history SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE`,
        userID,
    )
    if err != nil {
        return err
    }
    if _, err := res.RowsAffected(); err != nil {
        return err
    }
    return nil
}

func (r *Repository) DeleteNotification(ctx context.Context, notificationID string) error {
    res, err := r.db.ExecContext(
        ctx,
        `DELETE FROM notification_history WHERE id = $1`,
        notificationID,
    )
    if err != nil {
        return err
    }
    if rows, _ := res.RowsAffected(); rows == 0 {
        return sql.ErrNoRows
    }
    return nil
}

func (r *Repository) ClearHistory(ctx context.Context, userID string) error {
    _, err := r.db.ExecContext(
        ctx,
        `DELETE FROM notification_history WHERE user_id = $1`,
        userID,
    )
    return err
}