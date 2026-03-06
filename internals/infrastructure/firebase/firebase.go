package firebase

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/messaging"
)

func InitAuth() (*auth.Client, error) {
	ctx := context.Background()

	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting firebase auth client: %v", err)
	}

	fmt.Println("Firebase Admin SDK Initialized Successfully")
	return authClient, nil
}

func InitFCM() (*messaging.Client, error) {
    ctx := context.Background()

    app, err := firebase.NewApp(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("error initializing firebase app: %v", err)
    }

    fcmClient, err := app.Messaging(ctx)
    if err != nil {
        return nil, fmt.Errorf("error getting firebase messaging client: %v", err)
    }

    return fcmClient, nil
}