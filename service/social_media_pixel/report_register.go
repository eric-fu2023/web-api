package social_media_pixel

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"web-api/model"
	"web-api/util"
)

type ConfigDetails struct {
	SmPlatform int
	ID         string
	Token      string
}

const (
	SmPlatformTikTok   = 1
	SmPlatformFacebook = 2
)

func ReportRegisterConversion(ctx context.Context, user model.User) {
	configDetails, ok := Config[user.Channel]
	if !ok {
		return
	}

	if configDetails.SmPlatform == SmPlatformTikTok {
		err := reportRegisterTikTok(ctx, configDetails)
		if err != nil {
			util.GetLoggerEntry(ctx).Errorf("ReportRegisterConversionTikTok error: %s", err.Error())
			return
		}
	} else if configDetails.SmPlatform == SmPlatformFacebook {
		err := reportRegisterFacebook(ctx, configDetails)
		if err != nil {
			util.GetLoggerEntry(ctx).Errorf("ReportRegisterConversionFacebook error: %s", err.Error())
			return
		}
	}

	return
}

func reportRegisterTikTok(ctx context.Context, configDetails ConfigDetails) error {
	url := "https://business-api.tiktok.com/open_api/v1.3/event/track/"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{
    "event_source": "web",
    "event_source_id": "%s",
    "data": [
        {
            "event": "CompleteRegistration",
            "event_id": "CompleteRegistration",
            "event_time": %d,
            "properties": {
                "contents": [
                    {
                        "content_name": "CompleteRegistration"
                    }
                ]
            },
            "page": {
                "url": "aha888.vip"
            },
            "test_event_code": ""
        }
    ]
}`, configDetails.ID, time.Now().Unix()))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		util.GetLoggerEntry(ctx).Errorf("NewRequest error: %s", err.Error())
		return err
	}
	req.Header.Add("Access-Token", configDetails.Token)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		util.GetLoggerEntry(ctx).Errorf("client.Do error: %s", err.Error())
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		util.GetLoggerEntry(ctx).Errorf("ReadAll error: %s", err.Error())
		return err
	}

	return fmt.Errorf("unexpected response: %s", string(body))
}

func reportRegisterFacebook(ctx context.Context, configDetails ConfigDetails) error {
	url := fmt.Sprintf("https://graph.facebook.com/v19.0/%s/events", configDetails.ID)
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{
    "data": [
        {
            "action_source": "website",
            "event_name": "CompleteRegistration",
            "event_time": %d,
            "custom_data": {},
            "user_data": {
                "em": [""],
                "ph": []
            }
        }
    ],
    "access_token": "%s"
}`, time.Now().Unix(), configDetails.Token))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		util.GetLoggerEntry(ctx).Errorf("NewRequest error: %s", err.Error())
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		util.GetLoggerEntry(ctx).Errorf("client.Do error: %s", err.Error())
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		util.GetLoggerEntry(ctx).Errorf("ReadAll error: %s", err.Error())
		return err
	}

	return fmt.Errorf("unexpected response: %s", string(body))
}
