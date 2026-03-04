package party

import (
	"context"
	"errors"
	"fmt"
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

            if visibility == RoomVisibilityPublic {
                fmt.Println("Match! Attempting to write to Redis Lobby...")
                errLobby := s.redisRepo.PublishToLobby(ctx, &RoomSummary{
                    RoomID:          room.ID,
                    RoomName:        name,
                    HostUserID:      hostUserID,
                    HostDisplayName: "Host", // Pass actual display name
                    Visibility:      visibility,
                    // Note: Track info will be empty until plays a song
                })

                if errLobby != nil {
                    fmt.Println("REDIS PIPELINE ERROR:", errLobby)
                } else {
                    fmt.Println("Successfully written to Redis Lobby!")
                }
            }

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

    user, err := s.repo.GetUserSkinnyInfo(ctx, userID)
    if err != nil {
        return err
    }

    err = s.redisRepo.AddMember(ctx, roomID, userID)
    if err != nil {
        return err
    }

    count, _ := s.redisRepo.GetMemberCount(ctx, roomID)

    if room.Visibility == RoomVisibilityPublic {
        _ = s.redisRepo.SetLobbyCount(ctx, roomID, count)
    }

    event := map[string]interface{}{
        "type":           "USER_JOINED",
        "user_id": user.UserID,
        "username": user.Username,
        "display_name": user.DisplayName,
        "avatar_url": user.AvatarURL,
        "listener_count": count,
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
    _ = s.redisRepo.SetLobbyCount(ctx, roomID, count)
    
    event := map[string]interface{}{
        "type":           "USER_LEFT",
        "listener_count": count,
        "target_user_id": userID,
    }
    _ = s.redisRepo.PublishRoomEvent(ctx, roomID, event)
    return err
}

func (s *Service) ListActiveRooms(ctx context.Context, limit, offset int64) ([]RoomSummary, int64, error) {
    summaries, err := s.redisRepo.GetLobbyFeed(ctx, limit, offset)
    if err != nil {
        return nil, 0, err
    }

    nextOffset := int64(0)
    if int64(len(summaries)) == limit {
        nextOffset = offset + limit
    }

    return summaries, nextOffset, nil
}

func (s *Service) GetListeners(ctx context.Context, roomID string) ([]ListenerInfo, error) {
    userIDs, err := s.redisRepo.client.SMembers(ctx, roomMembersKey(roomID)).Result()
    if err != nil {
        return nil, err
    }
    if len(userIDs) == 0 {
        return []ListenerInfo{}, nil
    }

    users, err := s.repo.GetListeners(ctx, userIDs)
    if err != nil {
        return nil, err
    }

    listeners := make([]ListenerInfo, 0, len(users))
    for _, u := range users {
        listeners = append(listeners, ListenerInfo{
            UserID:      u.UserID,
            Username:    u.Username,
            DisplayName: u.DisplayName,
            AvatarURL:   u.AvatarURL,
        })
    }
    return listeners, nil
}

func (s *Service) SendLiveChatMessage(ctx context.Context, roomID, userID, text string) error {
    if roomID == "" || userID == "" || text == "" {
        return ErrInvalidInput
    }

    event := map[string]interface{}{
        "type":    "LIVE_CHAT_MESSAGE",
        "user_id": userID,
        "text": text,
    }

    return s.redisRepo.PublishRoomEvent(ctx, roomID, event)
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