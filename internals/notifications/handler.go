package notifications

import (
    "context"

    pb "github.com/dis70rt/bluppi-backend/internals/gen/notifications"
    "github.com/dis70rt/bluppi-backend/internals/infrastructure/middlewares"
    "google.golang.org/protobuf/types/known/emptypb"
)

type GrpcHandler struct {
    pb.UnimplementedNotificationServiceServer
    service *Service
}

func NewGrpcHandler(s *Service) *GrpcHandler {
    return &GrpcHandler{service: s}
}

func (h *GrpcHandler) RegisterDevice(ctx context.Context, req *pb.RegisterDeviceRequest) (*emptypb.Empty, error) {
    userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, err
    }

    deviceType := ""
    switch req.DeviceType {
    case pb.DeviceType_DEVICE_TYPE_ANDROID:
        deviceType = "android"
    case pb.DeviceType_DEVICE_TYPE_IOS:
        deviceType = "ios"
    default:
        deviceType = ""
    }

    err = h.service.RegisterDevice(ctx, userID, req.FcmToken, deviceType)
    if err != nil {
        return nil, mapError(err)
    }
    return &emptypb.Empty{}, nil
}

func (h *GrpcHandler) UnregisterDevice(ctx context.Context, req *pb.UnregisterDeviceRequest) (*emptypb.Empty, error) {
    err := h.service.UnregisterDevice(ctx, req.FcmToken)
    if err != nil {
        return nil, mapError(err)
    }
    return &emptypb.Empty{}, nil
}

func (h *GrpcHandler) GetPreferences(ctx context.Context, _ *emptypb.Empty) (*pb.NotificationPreferences, error) {
    userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, err
    }

    prefs, err := h.service.GetPreferences(ctx, userID)
    if err != nil {
        return nil, mapError(err)
    }
    return mapPreferencesToProto(prefs), nil
}

func (h *GrpcHandler) UpdatePreferences(ctx context.Context, req *pb.UpdatePreferencesRequest) (*emptypb.Empty, error) {
    userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, err
    }

    prefs := req.Preferences
    if prefs == nil {
        return nil, mapError(ErrInvalidInput)
    }

    err = h.service.UpdatePreferences(ctx, userID, NotificationPreferences{
        PushNotificationsEnabled:  prefs.PushNotificationsEnabled,
        EmailNotificationsEnabled: prefs.EmailNotificationsEnabled,
        PartyInvitesEnabled:       prefs.PartyInvitesEnabled,
        NewFollowersEnabled:       prefs.NewFollowersEnabled,
        FollowRequestEnabled:      prefs.FollowRequestEnabled,
        FollowerListeningEnabled:  prefs.FollowerListeningEnabled,
    })
    if err != nil {
        return nil, mapError(err)
    }
    return &emptypb.Empty{}, nil
}

func (h *GrpcHandler) GetHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
    userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, err
    }

    list, unread, err := h.service.GetHistory(ctx, userID, int(req.Limit), int(req.Offset))
    if err != nil {
        return nil, mapError(err)
    }

    return &pb.GetHistoryResponse{
        Notifications: mapNotificationsToProto(list),
        TotalUnread:   int32(unread),
    }, nil
}

func (h *GrpcHandler) MarkAsRead(ctx context.Context, req *pb.MarkAsReadRequest) (*emptypb.Empty, error) {
    err := h.service.MarkAsRead(ctx, req.NotificationIds)
    if err != nil {
        return nil, mapError(err)
    }
    return &emptypb.Empty{}, nil
}

func (h *GrpcHandler) DeleteNotification(ctx context.Context, req *pb.DeleteNotificationRequest) (*emptypb.Empty, error) {
    err := h.service.DeleteNotification(ctx, req.NotificationId)
    if err != nil {
        return nil, mapError(err)
    }
    return &emptypb.Empty{}, nil
}

func (h *GrpcHandler) ClearHistory(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
    userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, err
    }

    err = h.service.ClearHistory(ctx, userID)
    if err != nil {
        return nil, mapError(err)
    }
    return &emptypb.Empty{}, nil
}

func (h *GrpcHandler) GetUnreadCount(ctx context.Context, _ *emptypb.Empty) (*pb.GetUnreadCountResponse, error) {
    userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, err
    }

    count, err := h.service.GetUnreadCount(ctx, userID)
    if err != nil {
        return nil, mapError(err)
    }
    return &pb.GetUnreadCountResponse{Count: int32(count)}, nil
}