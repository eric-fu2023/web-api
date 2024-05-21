package social_media_pixel

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strconv"
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

// deprecated: FE will do the reporting instead
func ReportRegisterConversion(ctx context.Context, user model.User) {

	configDetails, ok := Config[user.Channel]
	if !ok {
		return
	}
	fmt.Printf("Debug789: ReportRegisterConversion, channel: %s\n", user.Channel)

	if configDetails.SmPlatform == SmPlatformTikTok {
		err := reportRegisterTikTok(ctx, configDetails, user)
		if err != nil {
			util.GetLoggerEntry(ctx).Errorf("ReportRegisterConversionTikTok error: %s", err.Error())
			return
		}
	} else if configDetails.SmPlatform == SmPlatformFacebook {
		err := reportRegisterFacebook(ctx, configDetails, user)
		if err != nil {
			util.GetLoggerEntry(ctx).Errorf("ReportRegisterConversionFacebook error: %s", err.Error())
			return
		}
	}

	return
}

func reportRegisterTikTok(ctx context.Context, configDetails ConfigDetails, user model.User) error {
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

func reportRegisterFacebook(ctx context.Context, configDetails ConfigDetails, user model.User) error {
	fmt.Println("Debug789: reportRegisterFacebook")
	url := fmt.Sprintf("https://graph.facebook.com/v19.0/%s/events", configDetails.ID)
	method := "POST"

	h := sha256.New()
	h.Write([]byte(strconv.FormatInt(user.ID, 10)))
	userIdHash := hex.EncodeToString(h.Sum(nil))

	// generate uuid for event id
	eventId := uuid.NewString()

	payload := strings.NewReader(fmt.Sprintf(`{
    "data": [
        {
            "action_source": "website",
            "event_name": "CompleteRegistration",
            "event_time": %d,
            "custom_data": {},
            "user_data": {
                "external_id": "%s",
				"client_ip_address": "%s",
				"client_user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
            },
			"event_id": "%s"
        }
    ],
    "access_token": "%s"
}`, time.Now().Unix(), userIdHash, user.RegistrationIp, eventId, configDetails.Token))

	fmt.Printf("Debug789: userIdHash: %s, user.RegistrationIp: %s, eventId: %s, configDetails.Token: %s\n", userIdHash, user.RegistrationIp, eventId, configDetails.Token)

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

	//if res.StatusCode == http.StatusOK {
	//	return nil
	//}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		util.GetLoggerEntry(ctx).Errorf("ReadAll error: %s", err.Error())
		return err
	}

	fmt.Printf("Debug963: ResponseValue: %s\n", string(body))

	if res.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("unexpected response: %s", string(body))
}
