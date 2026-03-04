package party

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type Querier interface {
	sqlx.QueryerContext
	sqlx.ExecerContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Rebind(query string) string
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

func baseRoomSelect() string {
	return `
		SELECT id, name, code, status, visibility, invite_only,
		       host_user_id, created_at, updated_at
		FROM rooms
	`
}

func (r *Repository) CreateRoom(ctx context.Context, room *Room) (*Room, error) {
	var created Room

	err := r.db.GetContext(ctx, &created, `
		INSERT INTO rooms (
			id, name, code, status, visibility, invite_only, host_user_id
		)
		VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6
		)
		RETURNING id, name, code, status, visibility, invite_only,
		          host_user_id, created_at, updated_at
	`,
		room.Name,
		room.Code,
		room.Status,
		room.Visibility,
		room.InviteOnly,
		room.HostUserID,
	)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *Repository) EndHostActiveRooms(ctx context.Context, hostUserID string) error {
    _, err := r.db.ExecContext(ctx, `
        UPDATE rooms 
        SET status = 'ENDED' 
        WHERE host_user_id = $1 AND status = 'ACTIVE'
    `, hostUserID)
    
    return err
}

func (r *Repository) EndRoom(ctx context.Context, roomID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE rooms
		SET status = 'ENDED'
		WHERE id = $1
	`, roomID)

	return err
}

func (r *Repository) GetRoom(ctx context.Context, roomID string) (*Room, error) {
	var room Room

	err := r.db.GetContext(ctx, &room, baseRoomSelect()+` WHERE id = $1`, roomID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &room, err
}

func (r *Repository) GetListeners(ctx context.Context, userIDs []string) ([]ListenerInfo, error) {
    if len(userIDs) == 0 {
        return []ListenerInfo{}, nil
    }

    query, args, err := sqlx.In(`
        SELECT id, username, name, profile_pic
        FROM users
        WHERE id IN (?)
    `, userIDs)
    if err != nil {
        return nil, err
    }
    query = r.db.Rebind(query)

    var listeners []ListenerInfo
    if err := r.db.SelectContext(ctx, &listeners, query, args...); err != nil {
        return nil, err
    }
    return listeners, nil
}

func (r *Repository) GetUserSkinnyInfo(ctx context.Context, userID string) (*ListenerInfo, error) {
	var listener ListenerInfo
    err := r.db.GetContext(ctx, &listener, `
        SELECT id AS id, username, name, profile_pic
        FROM users
        WHERE id = $1
    `, userID)
    if err != nil {
        return nil, err
    }
    return &listener, nil
}