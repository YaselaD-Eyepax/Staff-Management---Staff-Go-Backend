package workers

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var fcmClient *messaging.Client

func InitFCM(credPath string) (*messaging.Client, error) {
    if fcmClient != nil {
        return fcmClient, nil
    }

    opt := option.WithCredentialsFile(credPath)

    app, err := firebase.NewApp(context.Background(), nil, opt)
    if err != nil {
        return nil, err
    }

    client, err := app.Messaging(context.Background())
    if err != nil {
        return nil, err
    }

    log.Println("FCM client initialized ✔")

    fcmClient = client
    return client, nil
}
