package activity

import (
    "context"
)

type Service struct {
    graphRepo *GraphRepository
    sqlRepo   *Repository
}

func NewService(graphRepo *GraphRepository, sqlRepo *Repository) *Service {
    return &Service{
        graphRepo: graphRepo,
        sqlRepo:   sqlRepo,
    }
}

func (s *Service) GetFriendsFeed(ctx context.Context, userID string, limit, offset int32) ([]Activity, error) {
    if userID == "" {
        return nil, ErrInvalidInput
    }

    // 1. Get raw graph data
    feed, err := s.graphRepo.GetSortedFriendsFeed(ctx, userID, limit, offset)
    if err != nil {
        return nil, err
    }

    if len(feed) == 0 {
        return []Activity{}, nil
    }

    // 2. Extract IDs for hydration
    var userIDs []string
    var trackIDs []string
    for _, f := range feed {
        userIDs = append(userIDs, f.FriendID)
        if f.TrackID != nil {
            trackIDs = append(trackIDs, *f.TrackID)
        }
    }

    // 3. Batch fetch metadata from SQL safely
    userMap, err := s.sqlRepo.GetUsersByIDs(ctx, userIDs)
    if err != nil {
        return nil, err
    }

    trackMap, err := s.sqlRepo.GetTracksByIDs(ctx, trackIDs)
    if err != nil {
        return nil, err
    }

    // 4. Stitch everything together
    var activities []Activity
    for _, f := range feed {
        uData, userOk := userMap[f.FriendID]
        if !userOk {
            continue // Ghost user, skip
        }

        act := Activity{
            FriendID:   f.FriendID,
            FriendName: uData.Name,
            FriendUsername: uData.Username,
            Status:     f.Status,
            LastSeen:   f.LastActive, // Bind graph's Last Seen
        }

        if uData.ProfilePic != nil {
            act.FriendAvatarURL = *uData.ProfilePic
        }

        if f.TrackID != nil {
            if tData, trackOk := trackMap[*f.TrackID]; trackOk {
                act.TrackID = &tData.TrackID
                act.TrackTitle = &tData.Title
                act.TrackArtist = &tData.Artists
                if tData.ImageSmall != nil {
                    act.TrackCoverURL = tData.ImageSmall
                }
                if tData.PreviewURL != nil {
                    act.TrackPreviewURL = tData.PreviewURL // Bind Preview URL
                }
            }
        }

        activities = append(activities, act)
    }

    return activities, nil
}