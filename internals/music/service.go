package music

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
    ErrTrackNotFound = errors.New("track not found")
    ErrInvalidInput  = errors.New("invalid input")
    ErrAlreadyLiked  = errors.New("track already liked")
    ErrNotLiked      = errors.New("track not liked")
    ErrHistoryEmpty  = errors.New("history is empty")
)

type Service struct {
    repo *Repository
    graphRepo *GraphRepository
}

func NewService(repo *Repository, graphRepo *GraphRepository) *Service {
    return &Service{repo: repo, graphRepo: graphRepo}
}

// ----------------- Core Track Reading -----------------

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

// ----------------- Search & Discovery -----------------

func (s *Service) SearchTracks(
    ctx context.Context,
    query string,
    limit int,
    cursor string,
) ([]SearchTrack, string, error) {

    if limit <= 0 || limit > 100 {
        limit = 20
    }

    return s.repo.SearchTracks(ctx, query, limit, cursor)
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

    err := s.repo.LikeTrack(ctx, userID, trackID)
    if err != nil {
        return err
    }

    err = s.graphRepo.LikeTrack(ctx, userID, trackID)
    if err != nil {
        return err
    }

    return nil
}

func (s *Service) UnlikeTrack(ctx context.Context, userID, trackID string) error {
    if userID == "" || trackID == "" {
        return ErrInvalidInput
    }

    err := s.repo.UnlikeTrack(ctx, userID, trackID)
    if errors.Is(err, sql.ErrNoRows) {
        return ErrNotLiked
    } else if err != nil {
        return err
    }

    err = s.graphRepo.UnlikeTrack(ctx, userID, trackID)
    if err != nil {
        return err
    }

    return nil
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
    err := s.repo.AddTrackToHistory(ctx, userID, trackID)
    if err != nil {
        return err
    }

    err = s.graphRepo.LogListen(ctx, userID, trackID)
    if err != nil {
        return err
    }

    return nil
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

func (s *Service) WeeklyDiscoverTracks(ctx context.Context, userID string, limit int) ([]Track, error) {
    if userID == "" {
        return nil, ErrInvalidInput
    }
    if limit <= 0 || limit > 50 {
        limit = 20
    }

    var finalTrackIDs []string
    seen := make(map[string]bool)

    sevenDaysAgo := time.Now().AddDate(0, 0, -7).Unix()
    recentIDs, err := s.graphRepo.GetWeeklyDiscover(ctx, userID, limit, sevenDaysAgo)
    if err == nil {
        for _, id := range recentIDs {
            finalTrackIDs = append(finalTrackIDs, id)
            seen[id] = true
        }
    }

    if len(finalTrackIDs) < limit {
        allTimeLimit := limit - len(finalTrackIDs)
        allTimeIDs, err := s.graphRepo.GetWeeklyDiscover(ctx, userID, allTimeLimit, 0)
        if err == nil {
            for _, id := range allTimeIDs {
                if !seen[id] {
                    finalTrackIDs = append(finalTrackIDs, id)
                    seen[id] = true
                }
            }
        }
    }

    var discoveredTracks []Track

    if len(finalTrackIDs) > 0 {
        tracks, err := s.repo.GetTracksByIDs(ctx, finalTrackIDs)
        if err == nil {
            trackMap := make(map[string]Track)
            for _, t := range tracks {
                trackMap[t.ID] = t
            }
            for _, id := range finalTrackIDs {
                if track, exists := trackMap[id]; exists {
                    discoveredTracks = append(discoveredTracks, track)
                }
            }
        }
    }

    // Absolute Fallback
    if len(discoveredTracks) < limit {
        remaining := limit - len(discoveredTracks)
        
        unseenPopularTracks, err := s.repo.GetUnseenPopularTracks(ctx, userID, remaining)
        if err == nil {
            for _, pt := range unseenPopularTracks {
                if !seen[pt.ID] {
                    discoveredTracks = append(discoveredTracks, pt)
                    seen[pt.ID] = true
                }
            }
        }
    }

    return discoveredTracks, nil
}