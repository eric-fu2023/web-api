package util

import (
	"context"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	"os"
)

var FCMFactory client

func InitFCMFactory() {
	opt := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_FIREBASE"))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic(err)
	}
	FCMFactory = client{
		app: app,
	}
}

type client struct {
	app    *firebase.App
	dryRun bool
}

func (c *client) NewClient(dryRun bool) client {
	return client{
		app:    c.app,
		dryRun: dryRun,
	}
}

func (c *client) SendMessageToAll(data map[string]string, notification messaging.Notification, fcmTokens []string) error {
	ctx := context.TODO()
	msgClient, err := c.app.Messaging(ctx)
	if err != nil {
		fmt.Println("init firebase messaging client error:", err.Error())
		return err
	}

	message := &messaging.MulticastMessage{
		Tokens:       fcmTokens,
		Data:         data,
		Notification: &notification,
	}

	if c.dryRun {
		resp, err := msgClient.SendEachForMulticastDryRun(ctx, message)
		if err != nil {
			return err
		}
		if resp.FailureCount > 0 {
			return errors.Errorf("failed sending %d messages\n", resp.FailureCount)
		}
	} else {
		resp, err := msgClient.SendEachForMulticast(ctx, message)
		if err != nil {
			return err
		}
		if resp.FailureCount > 0 {
			return errors.Errorf("failed sending %d messages\n", resp.FailureCount)
		}
	}

	return nil
}
