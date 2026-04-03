package tests

import (
	"context"

	eventbus "github.com/dis70rt/bluppi-backend/internals/infrastructure/eventBus"
	"google.golang.org/protobuf/proto"
)

type noOpPublisher struct{}

func (n *noOpPublisher) Publish(ctx context.Context, topic eventbus.EventType, payload proto.Message) error {
    return nil
}