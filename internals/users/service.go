package users

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrAlreadyFollowing   = errors.New("already following user")
	ErrNotFollowing       = errors.New("not following user")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ----- Core CRUD Operations -----

func (s *Service) CreateUser(ctx context.Context, u *User) error {
	if u == nil {
		return ErrInvalidInput
	}
	if u.ID == "" || u.Username == "" || u.Email == "" || u.Name == "" {
		return ErrInvalidInput
	}

	if exists, _ := s.repo.EmailExists(ctx, u.Email); exists {
		return ErrUserAlreadyExists
	}

	return s.repo.CreateUser(ctx, u)
}

func (s *Service) GetUserByID(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}

	u, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *Service) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	if username == "" {
		return nil, ErrInvalidInput
	}

	u, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *Service) UpdateUser(
	ctx context.Context,
	id string,
	update map[string]any,
) error {

	if id == "" || len(update) == 0 {
		return ErrInvalidInput
	}

	err := s.repo.UpdateUser(ctx, id, update)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrUserNotFound
	}
	return err
}

func (s *Service) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidInput
	}

	err := s.repo.DeleteUser(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrUserNotFound
	}
	return err
}

// ----- Check and Search -----

func (s *Service) UsernameExists(ctx context.Context, username string) (bool, error) {
	if username == "" {
		return false, ErrInvalidInput
	}
	return s.repo.UsernameExists(ctx, username)
}

func (s *Service) EmailExists(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, ErrInvalidInput
	}
	return s.repo.EmailExists(ctx, email)
}

func (s *Service) SearchUsers(
	ctx context.Context,
	query string,
	limit, offset int,
) ([]UserSearchResult, int, error) {

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	return s.repo.SearchUsers(ctx, query, limit, offset)
}

// ----- Follow Operations -----

func (s *Service) Follow(ctx context.Context, followerID, followeeID string) error {
	if followerID == "" || followeeID == "" {
		return ErrInvalidInput
	}
	if followerID == followeeID {
		return ErrInvalidInput
	}

	err := s.repo.Follow(ctx, followerID, followeeID)
	return err
}

func (s *Service) Unfollow(ctx context.Context, followerID, followeeID string) error {
	if followerID == "" || followeeID == "" {
		return ErrInvalidInput
	}

	err := s.repo.Unfollow(ctx, followerID, followeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFollowing
	}
	return err
}


// ----- Stats -----

func (s *Service) GetUserStats(ctx context.Context, userID string) (*UserStats, error) {
	if userID == "" {
		return nil, ErrInvalidInput
	}

	return s.repo.GetUserStats(ctx, userID)
}

func (s *Service) AddRecentSearch(ctx context.Context, userID, query string) error {
	if userID == "" || query == "" {
		return ErrInvalidInput
	}
	return s.repo.AddRecentSearch(ctx, userID, query)
}

func (s *Service) GetRecentSearches(
	ctx context.Context,
	userID string,
	limit int,
) ([]RecentSearch, error) {

	if userID == "" {
		return nil, ErrInvalidInput
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	return s.repo.GetRecentSearches(ctx, userID, limit)
}
