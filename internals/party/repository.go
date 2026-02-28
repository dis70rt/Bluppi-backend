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

// func (r *Repository) DeleteRoom(ctx context.Context, roomID string) error {
// 	res, err := r.db.ExecContext(ctx, `DELETE FROM rooms WHERE id = $1`, roomID)
// 	if err != nil {
// 		return err
// 	}

// 	if rows, _ := res.RowsAffected(); rows == 0 {
// 		return sql.ErrNoRows
// 	}

// 	return nil
// }

// func (r *Repository) ListRooms(
//     ctx context.Context,
//     visibility RoomVisibility,
//     limit, offset int,
// ) ([]RoomSummary, error) {

//     summaries := []RoomSummary{}

//     err := r.db.SelectContext(ctx, &summaries, `
//         SELECT
//             id   AS room_id,
//             name AS room_name,
//             host_user_id,
//             visibility,
//             (status = 'ACTIVE') AS is_live,
//             0 AS listener_count -- We will fill this in via Redis later

//         FROM rooms
//         WHERE ($1 = 'ROOM_VISIBILITY_UNSPECIFIED' OR visibility = $1)
//           AND status = 'ACTIVE' -- You probably only want to list ACTIVE rooms!
//         ORDER BY created_at DESC
//         LIMIT $2 OFFSET $3
//     `,
//         visibility,
//         limit,
//         offset,
//     )

//     return summaries, err
// }