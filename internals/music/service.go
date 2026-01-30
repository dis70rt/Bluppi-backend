package music

import (
    "context"
    "database/sql"
    "errors"
)

var (
    ErrTrackNotFound    = errors.New("track not found")
    ErrInvalidInput     = errors.New("invalid input")
    ErrAlreadyLiked     = errors.New("track already liked")
    ErrNotLiked         = errors.New("track not liked")
    ErrHistoryEmpty     = errors.New("history is empty")
)

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

// ----------------- Core Track CRUD -----------------

func (s *Service) CreateTrack(ctx context.Context, t *Track) error {
    if t == nil {
        return ErrInvalidInput
    }
    if t.ID == "" || t.Title == "" || t.Artist == "" {
        return ErrInvalidInput
    }
    
    return s.repo.CreateTrack(ctx, t)
}

func (s *Service) GetTrack(ctx context.Context, id string) (*Track, error) {
    if id == "" {
        return nil, ErrInvalidInput
    }

    t, err := s.repo.GetTrack(ctx, id)
    if err != nil {
        return nil, err
    }
    if t == nil {
        return nil, ErrTrackNotFound
    }

    return t, nil
}

func (s *Service) UpdateTrack(ctx context.Context, id string, fields map[string]any) error {
    if id == "" || len(fields) == 0 {
        return ErrInvalidInput
    }

    err := s.repo.UpdateTrack(ctx, id, fields)
    if errors.Is(err, sql.ErrNoRows) {
        return ErrTrackNotFound
    }
    return err
}

func (s *Service) DeleteTrack(ctx context.Context, id string) error {
    if id == "" {
        return ErrInvalidInput
    }

    err := s.repo.DeleteTrack(ctx, id)
    if errors.Is(err, sql.ErrNoRows) {
        return ErrTrackNotFound
    }
    return err
}

// ----------------- Search & Discovery -----------------

// // TODO: CHECK HERE AND THINK FURTHER IMPLEMENTATION. (It should be python code as go doesn't have the required package yet.)
func (s *Service) SearchTracks(
    ctx context.Context,
    query string,
    limit, offset int,
) ([]Track, int, error) {
    // Allow empty query to return generic results if desired, or validation:
    // if query == "" { return nil, 0, ErrInvalidInput }

    if limit <= 0 || limit > 100 {
        limit = 20
    }
    if offset < 0 {
        offset = 0
    }

    return s.repo.SearchTracks(ctx, query, limit, offset)
}

func (s *Service) GetPopularTracks(ctx context.Context, limit int) ([]Track, error) {
    if limit <= 0 || limit > 100 {
        limit = 20
    }
    return s.repo.GetPopularTracks(ctx, limit)
}

func (s *Service) GetTracksByGenre(
    ctx context.Context,
    genre string,
    limit, offset int,
) ([]Track, int, error) {
    if genre == "" {
        return nil, 0, ErrInvalidInput
    }

    if limit <= 0 || limit > 100 {
        limit = 20
    }
    if offset < 0 {
        offset = 0
    }

    return s.repo.GetTracksByGenre(ctx, genre, limit, offset)
}

// ----------------- User Interactions (Likes) -----------------

func (s *Service) LikeTrack(ctx context.Context, userID, trackID string) error {
    if userID == "" || trackID == "" {
        return ErrInvalidInput
    }

    return s.repo.LikeTrack(ctx, userID, trackID)
}

func (s *Service) UnlikeTrack(ctx context.Context, userID, trackID string) error {
    if userID == "" || trackID == "" {
        return ErrInvalidInput
    }

    err := s.repo.UnlikeTrack(ctx, userID, trackID)
    if errors.Is(err, sql.ErrNoRows) {
        return ErrNotLiked
    }
    return err
}

func (s *Service) IsTrackLiked(ctx context.Context, userID, trackID string) (bool, error) {
    if userID == "" || trackID == "" {
        return false, ErrInvalidInput
    }
    return s.repo.IsTrackLiked(ctx, userID, trackID)
}

func (s *Service) GetLikedTracks(
    ctx context.Context,
    userID string,
    limit, offset int,
) ([]LikedTrackEntry, int, error) {
    if userID == "" {
        return nil, 0, ErrInvalidInput
    }

    if limit <= 0 || limit > 100 {
        limit = 20
    }
    if offset < 0 {
        offset = 0
    }

    return s.repo.GetLikedTracks(ctx, userID, limit, offset)
}

// ----------------- History -----------------

func (s *Service) AddTrackToHistory(ctx context.Context, userID, trackID string) error {
    if userID == "" || trackID == "" {
        return ErrInvalidInput
    }
    return s.repo.AddTrackToHistory(ctx, userID, trackID)
}

func (s *Service) GetTrackHistory(
    ctx context.Context,
    userID string,
    limit, offset int,
) ([]HistoryEntry, int, error) {
    if userID == "" {
        return nil, 0, ErrInvalidInput
    }

    if limit <= 0 || limit > 100 {
        limit = 20
    }
    if offset < 0 {
        offset = 0
    }

    return s.repo.GetTrackHistory(ctx, userID, limit, offset)
}

func (s *Service) ClearTrackHistory(ctx context.Context, userID string) error {
    if userID == "" {
        return ErrInvalidInput
    }
    return s.repo.ClearTrackHistory(ctx, userID)
}