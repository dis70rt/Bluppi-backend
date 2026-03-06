package eventbus

import (
    "context"

    "google.golang.org/protobuf/proto"
)

type EventType string

const (
    UserFollowedTopic      EventType = "USER_FOLLOWED"
    PartyInviteTopic       EventType = "PARTY_INVITE"
    FollowerListeningTopic EventType = "FOLLOWER_LISTENING"
)
type Event struct {
    Type    EventType
    Payload []byte
}

type EventHandler func(ctx context.Context, event Event) error

type Publisher interface {
    Publish(ctx context.Context, topic EventType, payload proto.Message) error
}

type Consumer interface {
    Consume(
        ctx context.Context,
        group string,
        consumerName string,
        topics []EventType,
        handler EventHandler,
    ) error
}

type EventBus interface {
    Publisher
    Consumer
    Close() error
}