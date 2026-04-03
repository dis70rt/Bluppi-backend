package notifications

import (
    "database/sql"

    pb "github.com/dis70rt/bluppi-backend/internals/gen/notifications"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/timestamppb"
    "encoding/json"
)

func mapPreferencesToProto(p *NotificationPreferences) *pb.NotificationPreferences {
    if p == nil {
        return nil
    }
    return &pb.NotificationPreferences{
        PushNotificationsEnabled:  p.PushNotificationsEnabled,
        EmailNotificationsEnabled: p.EmailNotificationsEnabled,
        PartyInvitesEnabled:       p.PartyInvitesEnabled,
        NewFollowersEnabled:       p.NewFollowersEnabled,
        FollowRequestEnabled:      p.FollowRequestEnabled,
        FollowerListeningEnabled:  p.FollowerListeningEnabled,
    }
}

func mapNotificationToProto(n *NotificationHistory) *pb.InAppNotification {
    actionDataStr := ""
    if n.ActionData != nil {
        b, _ := json.Marshal(n.ActionData)
        actionDataStr = string(b)
    }
    return &pb.InAppNotification{
        Id:        n.ID,
        Type:      string(n.Type),
        Title:     n.Title,
        Body:      n.Body,
        ActionData: actionDataStr,
        IsRead:    n.IsRead,
        CreatedAt: timestamppb.New(n.CreatedAt),
    }
}

func mapNotificationsToProto(list []NotificationHistory) []*pb.InAppNotification {
    out := make([]*pb.InAppNotification, len(list))
    for i := range list {
        out[i] = mapNotificationToProto(&list[i])
    }
    return out
}

func mapError(err error) error {
    switch err {
    case ErrNotificationNotFound:
        return status.Error(codes.NotFound, err.Error())
    case ErrDeviceNotFound:
        return status.Error(codes.NotFound, err.Error())
    case ErrPreferencesNotFound:
        return status.Error(codes.NotFound, err.Error())
    case ErrInvalidInput, ErrInvalidDeviceType, ErrNoNotificationIDs:
        return status.Error(codes.InvalidArgument, err.Error())
    case sql.ErrNoRows:
        return status.Error(codes.NotFound, "resource not found")
    default:
        return status.Error(codes.Internal, err.Error())
    }
}