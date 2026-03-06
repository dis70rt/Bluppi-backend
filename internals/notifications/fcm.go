package notifications

import (
    "context"
    "fmt"
    "log"

    "firebase.google.com/go/v4/messaging"
)

type FCMSender struct {
    client *messaging.Client
}

func NewFCMSender(client *messaging.Client) *FCMSender {
    return &FCMSender{client: client}
}

func (f *FCMSender) SendPush(ctx context.Context, tokens []string, title, body string, data map[string]string) error {
    if len(tokens) == 0 {
        return nil
    }

    message := &messaging.MulticastMessage{
        Tokens: tokens,
        Notification: &messaging.Notification{
            Title: title,
            Body:  body,
        },
        Data: data,
    }

    response, err := f.client.SendEachForMulticast(ctx, message)
    if err != nil {
        return fmt.Errorf("failed to send FCM multicast: %w", err)
    }

    if response.FailureCount > 0 {
        log.Printf("%d FCM messages failed to deliver successfully", response.FailureCount)
        // Optionally inspect response.Responses to remove invalid tokens from DB
    }

    return nil
}