package party

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/lib/pq"
)

var (
    ErrInvalidInput     = errors.New("invalid input")
    ErrRoomNotFound     = errors.New("room not found")
    ErrRoomCodeConflict = errors.New("room code collision")
    ErrAlreadyInRoom    = errors.New("user already in room")
    ErrNotInRoom        = errors.New("user not in room")
)

func CaptureServerReceiveUs() int64 {
    return time.Now().UnixMicro()
}

func CaptureServerSendUs() int64 {
    return time.Now().UnixMicro()
}

type Service struct {
    repo *Repository
    redisRepo *RedisRepository
}

func NewService(repo *Repository, redisRepo *RedisRepository) *Service {
    rand.Seed(time.Now().UnixNano())
    return &Service{
        repo: repo,
        redisRepo: redisRepo,
    }
}

const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateRoomCode(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = alphabet[rand.Intn(len(alphabet))]
    }
    return string(b)
}

func (s *Service) CreateRoom(
    ctx context.Context,
    name string,
    visibility RoomVisibility,
    inviteOnly bool,
    hostUserID string,
) (*Room, error) {
    if name == "" || hostUserID == "" {
        return nil, ErrInvalidInput
    }
    if visibility == "" {
        visibility = RoomVisibilityPublic
    }

    err := s.repo.EndHostActiveRooms(ctx, hostUserID)
    if err != nil {
        return nil, err
    }

    for i := 0; i < 5; i++ {
        code := generateRoomCode(7)
        room, err := s.repo.CreateRoom(ctx, &Room{
            Name:       name,
            Code:       code,
            Status:     RoomStatusActive,
            Visibility: visibility,
            InviteOnly: inviteOnly,
            HostUserID: hostUserID,
        })
        if err == nil {
            _ = s.redisRepo.SetRoomHost(ctx, room.ID, hostUserID)
            _ = s.redisRepo.RecordHeartbeat(ctx, room.ID)
            _ = s.redisRepo.AddMember(ctx, room.ID, hostUserID)
            return room, nil
        }
        if isUniqueViolation(err) {
            continue
        }
    }
    return nil, ErrRoomCodeConflict
}

func (s *Service) RecordHeartbeat(ctx context.Context, roomID, userID string) error {
    if roomID == "" || userID == "" {
        return ErrInvalidInput
    }

    hostID, err := s.redisRepo.GetRoomHost(ctx, roomID)
    if err != nil {
        return err
    }

    if userID == hostID {
        return s.redisRepo.RecordHeartbeat(ctx, roomID)
    }

    return nil
}

func (s *Service) GetRoom(ctx context.Context, roomID string) (*Room, error) {
    if roomID == "" {
        return nil, ErrInvalidInput
    }
    room, err := s.repo.GetRoom(ctx, roomID)
    if err != nil {
        return nil, err
    }
    if room == nil {
        return nil, ErrRoomNotFound
    }
    return room, nil
}

func (s *Service) JoinRoom(ctx context.Context, roomID, userID string) error {
    if roomID == "" || userID == "" {
        return ErrInvalidInput
    }

    room, err := s.repo.GetRoom(ctx, roomID)
    if err != nil {
        return err
    }
    if room == nil || room.Status != "ACTIVE" {
        return ErrRoomNotFound 
    }

    err = s.redisRepo.AddMember(ctx, roomID, userID)
    if err != nil {
        return err
    }

    count, _ := s.redisRepo.GetMemberCount(ctx, roomID)

    event := map[string]interface{}{
        "type":           "USER_JOINED",
        "listener_count": count,
        "target_user_id": userID,
    }
    _ = s.redisRepo.PublishRoomEvent(ctx, roomID, event)

    return nil
}

func (s *Service) LeaveRoom(ctx context.Context, roomID, userID string) error {
    if roomID == "" || userID == "" {
        return ErrInvalidInput
    }

    hostID, _ := s.redisRepo.GetRoomHost(ctx, roomID)

    if userID == hostID {
        _ = s.repo.EndRoom(ctx, roomID)    
        _ = s.redisRepo.CleanupRoom(ctx, roomID)
        return nil
    }

    err := s.redisRepo.RemoveMember(ctx, roomID, userID)

    count, _ := s.redisRepo.GetMemberCount(ctx, roomID)
    
    event := map[string]interface{}{
        "type":           "USER_LEFT",
        "listener_count": count,
        "target_user_id": userID,
    }
    _ = s.redisRepo.PublishRoomEvent(ctx, roomID, event)
    return err
}

func isUniqueViolation(err error) bool {
    if err == nil {
        return false
    }
    
    pqErr, ok := err.(*pq.Error)
    if ok {
        return pqErr.Code == "23505"
    }
    
    return false
}