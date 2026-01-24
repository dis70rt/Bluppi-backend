package users

import "time"

type User struct {
	ID             string    `db:"id"`
	Email          string    `db:"email"`
	Username       string    `db:"username"`
	Name           string    `db:"name"`
	Bio            *string   `db:"bio"`
	Country        *string   `db:"country"`
	Phone          *string   `db:"phone"`
	ProfilePic     *string   `db:"profile_pic"`
	FavoriteGenres []string  `db:"favorite_genres"`
	FollowerCount  int       `db:"follower_count"`
	FollowingCount int       `db:"following_count"`
	CreatedAt      time.Time `db:"created_at"`
}

type Follow struct {
	FollowerID string    `db:"follower_id"`
	FolloweeID string    `db:"followee_id"`
	CreatedAt  time.Time `db:"created_at"`
}

type RecentSearch struct {
	ID         int64     `db:"id"`
	UserID     string    `db:"user_id"`
	Query      string    `db:"query"`
	SearchedAt time.Time `db:"searched_at"`
}