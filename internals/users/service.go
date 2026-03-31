package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dis70rt/bluppi-backend/internals/gen/events"
	eventbus "github.com/dis70rt/bluppi-backend/internals/infrastructure/eventBus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrAlreadyFollowing   = errors.New("already following user")
	ErrNotFollowing       = errors.New("not following user")
	ErrUserExists         = errors.New("user already exists")
)

type Service struct {
	repo *Repository
	graphRepo *GraphRepository
	eventBus eventbus.Publisher
}

func NewService(repo *Repository, graphRepo *GraphRepository, eventBus eventbus.Publisher) *Service {
	return &Service{
		repo: repo,
		graphRepo: graphRepo,
		eventBus: eventBus,
	}
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

func (s *Service) UserExists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, ErrInvalidInput
	}
	return s.repo.UserExists(ctx, id)
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
	if err != nil {
		return err
	}

	err = s.graphRepo.Follow(ctx, followerID, followeeID)
    if err != nil {
        return err
    }

	follower, _ := s.repo.GetUserByID(ctx, followerID)
	event := &events.UserFollowedEvent{
		FollowerId: followerID,
		FollowerName: follower.Name,
		FollowerAvatar: *follower.ProfilePic,
		FolloweeId: followeeID,
		OccurredAt: timestamppb.Now(),
	}

	_ = s.eventBus.Publish(ctx, eventbus.UserFollowedTopic, event)

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

	err = s.graphRepo.Unfollow(ctx, followerID, followeeID)
    if err != nil {
        return err
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

func (s *Service) GetFollowers(ctx context.Context, userID, cursor string, limit int) ([]FollowEntry, string, error) {
    if userID == "" {
        return nil, "", ErrInvalidInput
    }
    if limit <= 0 || limit > 100 {
        limit = 20
    }
    return s.repo.GetFollowers(ctx, userID, cursor, limit)
}

func (s *Service) GetFollowing(ctx context.Context, userID, cursor string, limit int) ([]FollowEntry, string, error) {
    if userID == "" {
        return nil, "", ErrInvalidInput
    }
    if limit <= 0 || limit > 100 {
        limit = 20
    }
    return s.repo.GetFollowing(ctx, userID, cursor, limit)
}

func (s *Service) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
    if followerID == "" || followeeID == "" {
        return false, ErrInvalidInput
    }
    return s.repo.IsFollowing(ctx, followerID, followeeID)
}

func (s *Service) GetSuggestedUsers(ctx context.Context, userID string, limit int, cursor string) ([]*User, string, error) {
	if userID == "" {
		return nil, "", ErrInvalidInput
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	suggestedIDs, nextCursor, err := s.graphRepo.GetSuggestedUsers(ctx, userID, limit, cursor);
	if err != nil {
		return nil, "", err
	}
	
	if len(suggestedIDs) == 0 {
        return []*User{}, "", nil
    }

	unsortedUsers, err := s.repo.GetUsersByIDs(ctx, suggestedIDs)
    if err != nil {
        return nil, "", err
    }

	userMap := make(map[string]*User)
    for _, u := range unsortedUsers {
        userMap[u.ID] = u
    }

	var rankedUsers []*User
    for _, id := range suggestedIDs {
        if user, exists := userMap[id]; exists {
            rankedUsers = append(rankedUsers, user)
        }
    }

    return rankedUsers, nextCursor, nil
}