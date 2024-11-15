// Package util
// [FCM] deprecated. see common-functions
package util

import (
	"context"
	"fmt"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// FCMFactory Deprecated
var FCMFactory func(bool) client

// InitFCMFactory Deprecated
func InitFCMFactory(opt option.ClientOption) {
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic(err)
	}
	FCMFactory = func(dryRun bool) client {
		return client{
			app:    app,
			dryRun: dryRun,
		}
	}
}

// client Deprecated
type client struct {
	app    *firebase.App
	dryRun bool
}

// SendMessageToAll Deprecated

func (c *client) SendMessageToAll(ctx context.Context, data map[string]string, notification messaging.Notification, fcmTokens []string) error {
	msgClient, err := c.app.Messaging(ctx)
	if err != nil {
		return err
	}

	var grps [][]string
	var i, j int
	q := len(fcmTokens) / 500 // MulticastMessage can only send to 500 tokens
	r := len(fcmTokens) % 500
	for i = 0; i < q; i++ {
		j = i + 1
		grps = append(grps, fcmTokens[i*500:j*500])
	}
	if r > 0 {
		grps = append(grps, fcmTokens[j*500:])
	}

	var failures int
	errStrMap := map[string]struct{}{}
	for _, tokens := range grps {
		message := &messaging.MulticastMessage{
			Tokens:       tokens,
			Data:         data,
			Notification: &notification,
		}

		var resp *messaging.BatchResponse

		var cast func(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error)

		if c.dryRun {
			cast = msgClient.SendEachForMulticastDryRun
		} else {
			cast = msgClient.SendEachForMulticast
		}

		resp, err = cast(ctx, message)
		if err != nil {
			return err
		}
		if resp.FailureCount > 0 {
			failures += resp.FailureCount
			for _, respresp := range resp.Responses {
				if err := respresp.Error; err != nil {
					errStrMap[respresp.Error.Error()] = struct{}{}
				}
			}
		}
	}

	if failures > 0 {
		errrrrr := &strings.Builder{}
		for errStr := range errStrMap {
			errrrrr.WriteString(errStr)
			errrrrr.WriteString("\\")
		}
		return fmt.Errorf("failed sending %d messages. errors %s", failures, errrrrr.String())
	}

	return nil
}
