package notifications

import (
    "context"
    "log"

    "github.com/dis70rt/bluppi-backend/internals/gen/events"
    eventbus "github.com/dis70rt/bluppi-backend/internals/infrastructure/eventBus"
    "google.golang.org/protobuf/proto"
)

type Consumer struct {
    service 	*Service
    bus     	eventbus.Consumer
	fcmSender 	*FCMSender
}

func NewConsumer(service *Service, bus eventbus.Consumer, fcmSender *FCMSender) *Consumer {
    return &Consumer{
        service: service,
        bus:     bus,
		fcmSender: fcmSender,
    }
}

func (c *Consumer) Start(ctx context.Context) error {
    topics := []eventbus.EventType{
        eventbus.UserFollowedTopic,
        eventbus.PartyInviteTopic,
        eventbus.FollowerListeningTopic,
    }

    return c.bus.Consume(
        ctx,
        "notification-service-group",
        "worker-1",                   // Consumer name (could be hostname or UUID for multiple replicas)
        topics,
        c.handleEvent,
    )
}

func (c *Consumer) handleEvent(ctx context.Context, e eventbus.Event) error {
    switch e.Type {
    case eventbus.UserFollowedTopic:
        return c.handleUserFollowed(ctx, e.Payload)
    case eventbus.PartyInviteTopic:
        return c.handlePartyInvite(ctx, e.Payload)
    case eventbus.FollowerListeningTopic:
        return c.handleFollowerListening(ctx, e.Payload)
    default:
        log.Printf("ignored unknown event type: %s", e.Type)
        return nil
    }
}

func (c *Consumer) handleUserFollowed(ctx context.Context, payload []byte) error {
    var pb events.UserFollowedEvent
    if err := proto.Unmarshal(payload, &pb); err != nil {
        return err
    }

    actionData := ActionData{
        "follower_id":     pb.FollowerId,
        "follower_name":   pb.FollowerName,
        "follower_avatar": pb.FollowerAvatar,
    }

    err := c.service.CreateNotification(
        ctx,
        pb.FolloweeId,
        NotificationTypeNewFollower,
        pb.FollowerName+" started following you",
        "Tap to view their profile",
        actionData,
    )

    c.sendPushIfEnabled(
        ctx,
        pb.FolloweeId,
        pb.FollowerName+" started following you",
        "Tap to view their profile",
        map[string]string{
            "type":         string(NotificationTypeNewFollower),
            "follower_id":  pb.FollowerId,
            "follower_name": pb.FollowerName,
        },
    )

    return err
}

func (c *Consumer) handlePartyInvite(ctx context.Context, payload []byte) error {
    var pb events.PartyInviteEvent
    if err := proto.Unmarshal(payload, &pb); err != nil {
        return err
    }

    actionData := ActionData{
        "room_id":        pb.RoomId,
        "room_name":      pb.RoomName,
        "inviter_id":     pb.InviterId,
        "inviter_name":   pb.InviterName,
        "inviter_avatar": pb.InviterAvatar,
    }

    err := c.service.CreateNotification(
        ctx,
        pb.TargetUserId,
        NotificationTypePartyInvite,
        pb.InviterName+" invited you to a party!",
        "Join the room: "+pb.RoomName,
        actionData,
    )

    c.sendPushIfEnabled(
        ctx,
        pb.TargetUserId,
        pb.InviterName+" invited you to a party!",
        "Join the room: "+pb.RoomName,
        map[string]string{
            "type":      string(NotificationTypePartyInvite),
            "room_id":   pb.RoomId,
            "room_name": pb.RoomName,
        },
    )

    return err
}

func (c *Consumer) handleFollowerListening(ctx context.Context, payload []byte) error {
    var pb events.FollowerListeningEvent
    if err := proto.Unmarshal(payload, &pb); err != nil {
        return err
    }

    actionData := ActionData{
        "broadcaster_id":   pb.BroadcasterId,
        "broadcaster_name": pb.BroadcasterName,
        "room_id":          pb.RoomId,
        "track_name":       pb.TrackName,
    }

    for _, targetUserID := range pb.TargetFollowerIds {
        err := c.service.CreateNotification(
            ctx,
            targetUserID,
            NotificationTypeFollowerListening,
            pb.BroadcasterName+" is listening to music",
            "Track: "+pb.TrackName,
            actionData,
        )
        if err != nil {
            log.Printf("failed to save listener notification for user %s: %v", targetUserID, err)
            continue
        }

        c.sendPushIfEnabled(
            ctx,
            targetUserID,
            pb.BroadcasterName+" is listening to music",
            "Track: "+pb.TrackName,
            map[string]string{
                "type":             string(NotificationTypeFollowerListening),
                "broadcaster_id":   pb.BroadcasterId,
                "broadcaster_name": pb.BroadcasterName,
                "room_id":          pb.RoomId,
                "track_name":       pb.TrackName,
            },
        )
    }

    return nil
}

func (c *Consumer) sendPushIfEnabled(ctx context.Context, userID string, title, body string, eventData map[string]string) {
    prefs, err := c.service.GetPreferences(ctx, userID)
    if err != nil || prefs == nil || !prefs.PushNotificationsEnabled {
        return
    }
    
    devices, err := c.service.GetActiveDeviceTokensByUserID(ctx, userID)
    if err != nil || len(devices) == 0 {
        return
    }

    var tokens []string
    for _, d := range devices {
        tokens = append(tokens, d.FCMToken)
    }

    err = c.fcmSender.SendPush(ctx, tokens, title, body, eventData)
    if err != nil {
        log.Printf("fcm push error for user %s: %v", userID, err)
    }
}