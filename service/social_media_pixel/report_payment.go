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

type PaymentDetails struct {
	Currency     string
	Value        int64
	CashMethodId int64
}

func ReportPayment(ctx context.Context, user model.User, paymentDetails PaymentDetails) {
	configDetails, ok := Config[user.Channel]
	if !ok {
		return
	}

	fmt.Printf("Debug963: ReportPayment channel: %s, smPlatform: %d, id: %s, token: %s\n", user.Channel, configDetails.SmPlatform, configDetails.ID, configDetails.Token)
	fmt.Printf("Debug456 configDetails Platform: %d, ID: %s, Token: %s\n", configDetails.SmPlatform, configDetails.ID, configDetails.Token)

	if configDetails.SmPlatform == SmPlatformTikTok {
		fmt.Printf("Debug456: TikTok\n")
		err := reportPaymentTikTok(ctx, configDetails, paymentDetails)
		if err != nil {
			util.GetLoggerEntry(ctx).Errorf("ReportPaymentTikTok error: %s", err.Error())
			return
		}
	} else if configDetails.SmPlatform == SmPlatformFacebook {
		fmt.Printf("Debug456: Facebook\n")
		err := reportPaymentFacebook(ctx, configDetails, paymentDetails, user)
		if err != nil {
			util.GetLoggerEntry(ctx).Errorf("ReportPaymentFacebook error: %s", err.Error())
			return
		}
	}

	return
}

func reportPaymentTikTok(ctx context.Context, configDetails ConfigDetails, paymentDetails PaymentDetails) error {
	url := "https://business-api.tiktok.com/open_api/v1.3/event/track/"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{
    "event_source": "web",
    "event_source_id": "%s",
    "data": [
        {
            "event": "CompletePayment",
            "event_id": "CompletePayment",
            "event_time": %d,
            "properties": {
				"currency": "%s",
				"value": %.2f,	
                "contents": [
                    {
                        "content_name": "CompletePayment",
						"content_id": "%d"
                    }
                ]
            },
            "page": {
                "url": "aha888.vip"
            },
            "test_event_code": ""
        }
    ]
}`, configDetails.ID, time.Now().Unix(), paymentDetails.Currency, float64(paymentDetails.Value/100), paymentDetails.CashMethodId))

	fmt.Printf("Debug456 EventSourceId: %s, EventTime: %d, Currency: %s, Value: %.2f\n",
		configDetails.ID, time.Now().Unix(), paymentDetails.Currency, float64(paymentDetails.Value/100),
	)

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

	fmt.Printf("Debug456 StatusCode: %d\n", res.StatusCode)

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

func reportPaymentFacebook(ctx context.Context, configDetails ConfigDetails, paymentDetails PaymentDetails, user model.User) error {
	fmt.Printf("Debug963: reportPaymentFacebook\n")
	url := fmt.Sprintf("https://graph.facebook.com/v19.0/%s/events", configDetails.ID)
	method := "POST"

	h := sha256.New()
	h.Write([]byte(strconv.FormatInt(user.ID, 10)))
	userIdHash := hex.EncodeToString(h.Sum(nil))

	// generate uuid for event id
	eventId := uuid.NewString()

	ip := user.LastLoginIp
	if ip == "" {
		ip = user.RegistrationIp
	}

	payload := strings.NewReader(fmt.Sprintf(`{
    "data": [
        {
            "action_source": "website",
            "event_name": "Purchase",
            "event_time": %d,
            "custom_data": {
				"currency": "%s",
				"value": %.2f
			},
            "user_data": {
                "external_id": "%s",
				"client_ip_address": "%s",
				"client_user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
            },
			"event_id": "%s"
        }
    ],
    "access_token": "%s"
}`, time.Now().Unix(), paymentDetails.Currency, float64(paymentDetails.Value/100), userIdHash, ip, eventId, configDetails.Token))

	fmt.Printf("Debug963: paymentDetails.Currency: %s\n", paymentDetails.Currency)
	fmt.Printf("Debug963: float64(paymentDetails.Value/100): %.2f\n", float64(paymentDetails.Value/100))
	fmt.Printf("Debug963: userIdHash: %s\n", userIdHash)
	fmt.Printf("Debug963: ip: %s\n", ip)
	fmt.Printf("Debug963: eventId: %s\n", eventId)
	fmt.Printf("Debug963: configDetails.Token: %s\n", configDetails.Token)

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
