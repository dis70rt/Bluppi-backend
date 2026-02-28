package party

import (
    // "time"

    roompb "github.com/dis70rt/bluppi-backend/internals/gen/rooms"
    "google.golang.org/protobuf/types/known/timestamppb"
)

func mapRoomStatus(s RoomStatus) roompb.RoomStatus {
    switch s {
    case RoomStatusActive:
        return roompb.RoomStatus_ROOM_STATUS_ACTIVE
    case RoomStatusEnded:
        return roompb.RoomStatus_ROOM_STATUS_ENDED
    default:
        return roompb.RoomStatus_ROOM_STATUS_UNSPECIFIED
    }
}

func mapRoomVisibility(v RoomVisibility) roompb.RoomVisibility {
    switch v {
    case RoomVisibilityPublic:
        return roompb.RoomVisibility_ROOM_VISIBILITY_PUBLIC
    case RoomVisibilityPrivate:
        return roompb.RoomVisibility_ROOM_VISIBILITY_PRIVATE
    default:
        return roompb.RoomVisibility_ROOM_VISIBILITY_UNSPECIFIED
    }
}

func mapRoomMemberRole(r RoomMemberRole) roompb.RoomMemberRole {
    switch r {
    case RoomRoleHost:
        return roompb.RoomMemberRole_ROOM_MEMBER_ROLE_HOST
    case RoomRoleListener:
        return roompb.RoomMemberRole_ROOM_MEMBER_ROLE_LISTENER
    default:
        return roompb.RoomMemberRole_ROOM_MEMBER_ROLE_UNSPECIFIED
    }
}

func mapRoomToProto(r *Room) *roompb.Room {
    if r == nil {
        return nil
    }
    return &roompb.Room{
        Id:         r.ID,
        Name:       r.Name,
        Code:       r.Code,
        Status:     mapRoomStatus(r.Status),
        Visibility: mapRoomVisibility(r.Visibility),
        HostUserId: r.HostUserID,
        CreatedAt:  timestamppb.New(r.CreatedAt),
        UpdatedAt:  timestamppb.New(r.UpdatedAt),
    }
}

func mapRoomMemberToProto(m *RoomMember) *roompb.RoomMember {
    if m == nil {
        return nil
    }
    var leftAt *timestamppb.Timestamp
    if m.LeftAt != nil {
        leftAt = timestamppb.New(*m.LeftAt)
    }
    return &roompb.RoomMember{
        RoomId:   m.RoomID,
        UserId:   m.UserID,
        Role:     mapRoomMemberRole(m.Role),
        JoinedAt: timestamppb.New(m.JoinedAt),
        LeftAt:   leftAt,
    }
}

func mapRoomSummaryToProto(s *RoomSummary) *roompb.RoomSummary {
    if s == nil {
        return nil
    }
    return &roompb.RoomSummary{
        RoomId:        s.RoomID,
        RoomName:      s.RoomName,
        HostUserId:    s.HostUserID,
        ListenerCount: s.ListenerCount,
        IsLive:        s.IsLive,
        Visibility:    mapRoomVisibility(s.Visibility),
    }
}

func mapRoomSummariesToProto(list []RoomSummary) []*roompb.RoomSummary {
    out := make([]*roompb.RoomSummary, 0, len(list))
    for _, s := range list {
        out = append(out, mapRoomSummaryToProto(&s))
    }
    return out
}