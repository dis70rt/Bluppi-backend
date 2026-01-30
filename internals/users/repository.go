package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	// "github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	// pq "github.com/lib/pq"
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

// ----------------- Core CRUD Operations -----------------

func (r *Repository) CreateUser(ctx context.Context, u *User) error {
    query := `
        INSERT INTO users (id, email, username, name, bio, country, phone, profile_pic, favorite_genres)
        VALUES (:id, :email, :username, :name, :bio, :country, :phone, :profile_pic, :favorite_genres)
    `
    
    _, err := r.db.NamedExecContext(ctx, query, u)
    return err
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	var u User

	err := r.db.GetContext(
		ctx,
		&u,
		`SELECT * FROM users WHERE id = $1`,
		id,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var u User

	err := r.db.GetContext(
		ctx,
		&u,
		`SELECT * FROM users WHERE username = $1`,
		username,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}

func (r *Repository) UpdateUser(ctx context.Context, id string, fields map[string]any) error {
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

    args = append(args, id)
    query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(setClauses, ", "), i)

    res, err := r.db.ExecContext(ctx, query, args...)
    if err != nil {
        return err
    }

    rows, err := res.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return ErrUserNotFound
    }

    return nil
}

func (r *Repository) DeleteUser(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Username and Email existence checks

func (r *Repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.GetContext(
		ctx,
		&exists,
		`SELECT EXISTS (SELECT 1 FROM users WHERE username = $1)`,
		username,
	)
	return exists, err
}

func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.GetContext(
		ctx,
		&exists,
		`SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`,
		email,
	)
	return exists, err
}

// ----------------- Search History Operations -----------------

type UserSearchResult struct {
	ID            string  `db:"id"`
	Username      string  `db:"username"`
	Name          string  `db:"name"`
	ProfilePic    *string `db:"profile_pic"`
	FollowerCount int     `db:"follower_count"`
}

func (r *Repository) SearchUsers(
	ctx context.Context,
	query string,
	limit, offset int,
) ([]UserSearchResult, int, error) {

	search := "%" + query + "%"
	users := []UserSearchResult{}

	err := r.db.SelectContext(
		ctx,
		&users,
		`
		SELECT id, username, name, profile_pic, follower_count
		FROM users
		WHERE username ILIKE $1 OR name ILIKE $1
		ORDER BY follower_count DESC
		LIMIT $2 OFFSET $3
		`,
		search, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}

	var total int
	err = r.db.GetContext(
		ctx,
		&total,
		`
		SELECT COUNT(*) FROM users
		WHERE username ILIKE $1 OR name ILIKE $1
		`,
		search,
	)

	return users, total, err
}

// ----------------- Follow Operations -----------------

func (r *Repository) Follow(ctx context.Context, followerID, followeeID string) error {
	_, err := r.db.ExecContext(
		ctx,
		`
		INSERT INTO follows (follower_id, followee_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
		`,
		followerID, followeeID,
	)
	return err
}

func (r *Repository) Unfollow(ctx context.Context, followerID, followeeID string) error {
	res, err := r.db.ExecContext(
		ctx,
		`DELETE FROM follows WHERE follower_id = $1 AND followee_id = $2`,
		followerID, followeeID,
	)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ----------------- Recent Searches Operations -----------------

type UserStats struct {
	LikedTracks    int `db:"liked_tracks"`
	TotalPlays     int `db:"total_plays"`
	FollowingCount int `db:"following_count"`
	FollowersCount int `db:"followers_count"`
}

func (r *Repository) GetUserStats(ctx context.Context, userID string) (*UserStats, error) {
	var stats UserStats

	err := r.db.GetContext(
		ctx,
		&stats,
		`
		SELECT
			(SELECT COUNT(*) FROM user_track WHERE user_id = $1 AND interaction_type = 'liked') AS liked_tracks,
			(SELECT COUNT(*) FROM history_tracks WHERE user_id = $1) AS total_plays,
			(SELECT COUNT(*) FROM follows WHERE follower_id = $1) AS following_count,
			(SELECT COUNT(*) FROM follows WHERE followee_id = $1) AS followers_count
		`,
		userID,
	)

	return &stats, err
}

func (r *Repository) AddRecentSearch(ctx context.Context, userID, query string) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO recent_searches (user_id, query) VALUES ($1, $2)`,
		userID, query,
	)
	return err
}

func (r *Repository) GetRecentSearches(
	ctx context.Context,
	userID string,
	limit int,
) ([]RecentSearch, error) {

	searches := []RecentSearch{}

	err := r.db.SelectContext(
		ctx,
		&searches,
		`
		SELECT * FROM recent_searches
		WHERE user_id = $1
		ORDER BY searched_at DESC
		LIMIT $2
		`,
		userID, limit,
	)

	return searches, err
}

func (r *Repository) DeleteRecentSearch(ctx context.Context, userID string, searchID string) error {
    res, err := r.db.ExecContext(
        ctx,
        `DELETE FROM recent_searches WHERE user_id = $1 AND id = $2`,
        userID, searchID,
    )
    if err != nil {
        return err
    }
    if rows, _ := res.RowsAffected(); rows == 0 {
        return sql.ErrNoRows
    }
    return nil
}

// ----------------- Follow List Operations -----------------

type FollowEntry struct {
    ID         string     `db:"id"`
    Username   string     `db:"username"`
    Name       string     `db:"name"`
    ProfilePic *string    `db:"profile_pic"`
    FollowedAt time.Time  `db:"followed_at"`
}

func (r *Repository) GetFollowers(ctx context.Context, userID string, limit, offset int) ([]FollowEntry, int, error) {
    followers := []FollowEntry{}

    err := r.db.SelectContext(
        ctx,
        &followers,
        `
        SELECT u.id, u.username, u.name, u.profile_pic, f.created_at AS followed_at
        FROM follows f
        JOIN users u ON f.follower_id = u.id
        WHERE f.followee_id = $1
        ORDER BY f.created_at DESC
        LIMIT $2 OFFSET $3
        `,
        userID, limit, offset,
    )
    if err != nil {
        return nil, 0, err
    }

    var total int
    err = r.db.GetContext(
        ctx,
        &total,
        `SELECT COUNT(*) FROM follows WHERE followee_id = $1`,
        userID,
    )

    return followers, total, err
}

func (r *Repository) GetFollowing(ctx context.Context, userID string, limit, offset int) ([]FollowEntry, int, error) {
    following := []FollowEntry{}

    err := r.db.SelectContext(
        ctx,
        &following,
        `
        SELECT u.id, u.username, u.name, u.profile_pic, f.created_at AS followed_at
        FROM follows f
        JOIN users u ON f.followee_id = u.id
        WHERE f.follower_id = $1
        ORDER BY f.created_at DESC
        LIMIT $2 OFFSET $3
        `,
        userID, limit, offset,
    )
    if err != nil {
        return nil, 0, err
    }

    var total int
    err = r.db.GetContext(
        ctx,
        &total,
        `SELECT COUNT(*) FROM follows WHERE follower_id = $1`,
        userID,
    )

    return following, total, err
}

func (r *Repository) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
    var exists bool
    err := r.db.GetContext(
        ctx,
        &exists,
        `SELECT EXISTS (SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = $2)`,
        followerID, followeeID,
    )
    return exists, err
}
