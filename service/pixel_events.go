package service

import (
	"log"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
)

type PixelRequestBody struct {
	Data        []PixelRequestBodyData `form:"data" json:"data"`
	AccessToken string                 `form:"access_token" json:"access_token"`
}
type PixelRequestBodyData struct {
	EventName    string     `form:"event_name" json:"event_name"`
	EventTime    int64      `form:"event_time" json:"event_time"`
	UserData     UserData   `form:"user_data" json:"user_data"`
	AppData      AppData    `form:"app_data" json:"app_data"`
	CustomData   CustomData `form:"custom_data" json:"custom_data"`
	ActionSource string     `form:"action_source" json:"action_source"`
}
type UserData struct {
	Email           []string `json:"em"`
	Phone           []string `json:"ph"`
	ClientIpAddress string   `json:"client_ip_address"`
	ExternalId      int64      `json:"external_id"`
}
type AppData struct {
	AdvertiserTrackingEnabled  int      `json:"advertiser_tracking_enabled"`
	ApplicationTrackingEnabled int      `json:"application_tracking_enabled"`
	Extinfo                    []string `json:"extinfo"`
}
type CustomData struct {
	UserId   int64 `json:"user_id"`
	Currency string `json:"currency"`
	Value    int64    `json:"value"`
}

const PixelURl = "https://graph.facebook.com/v21.0/839510968259106/events"

func PixelInstallEvent(client_ip string) {
	var pixelRequestBody PixelRequestBody
	pixelRequestBody.AccessToken = os.Getenv("PIXEL_ACCESS_TOKEN")
	pixelRequestBody.Data = []PixelRequestBodyData{{
		EventName:    "ViewContent",
		EventTime:    time.Now().Unix(),
		UserData:     UserData{
			ClientIpAddress: client_ip,
		},
		AppData:      AppData{
			AdvertiserTrackingEnabled:0, 
			ApplicationTrackingEnabled:0,
			Extinfo:[]string{"a2", "com.batace", "1.0", "1.0.11", "13.4.1", "android", "En_US", "IST", "AT&T", "320", "540", "2", "2", "13", "8", "INR"},
		},
		ActionSource: "app",
	}}

	cl := resty.New()
	_, err := cl.R().SetBody(pixelRequestBody).Post(os.Getenv("PIXEL_END_POINT"))
	if err != nil {
		log.Printf("Pixel Install Api Call error for user %v", err.Error())
	}
}

func PixelRegisterEvent(user_id int64, client_ip string) {
	var pixelRequestBody PixelRequestBody
	pixelRequestBody.AccessToken = os.Getenv("PIXEL_ACCESS_TOKEN")
	pixelRequestBody.Data = []PixelRequestBodyData{{
		EventName:    "CompleteRegistration",
		EventTime:    time.Now().Unix(),
		UserData:     UserData{
			Phone:           []string{"hashed_phone_number"},
			ClientIpAddress: client_ip,
			ExternalId:      user_id,
		},
		AppData:      AppData{
			AdvertiserTrackingEnabled:0, 
			ApplicationTrackingEnabled:0,
			Extinfo:[]string{"a2", "com.batace", "1.0", "1.0.11", "13.4.1", "android", "En_US", "IST", "AT&T", "320", "540", "2", "2", "13", "8", "INR"},
		},
		CustomData:   CustomData{
			UserId: user_id,
			Currency: "INR",
			Value: 0,
		},
		ActionSource: "app",
	}}

	cl := resty.New()
	_, err := cl.R().SetBody(pixelRequestBody).Post(os.Getenv("PIXEL_END_POINT"))
	if err != nil {
		log.Printf("Pixel Register Api Call error for user %v, %v", user_id, err.Error())
	}
}

func PixelFTDEvent(user_id int64, client_ip string, deposit_amount int64) {
	var pixelRequestBody PixelRequestBody
	pixelRequestBody.AccessToken = os.Getenv("PIXEL_ACCESS_TOKEN")
	pixelRequestBody.Data = []PixelRequestBodyData{{
		EventName:    "Purchase",
		EventTime:    time.Now().Unix(),
		UserData:     UserData{
			Phone:           []string{"hashed_phone_number"},
			ClientIpAddress: client_ip,
			ExternalId:      user_id,
		},
		AppData:      AppData{
			AdvertiserTrackingEnabled:0, 
			ApplicationTrackingEnabled:0,
			Extinfo:[]string{"a2", "com.batace", "1.0", "1.0.11", "13.4.1", "android", "En_US", "IST", "AT&T", "320", "540", "2", "2", "13", "8", "INR"},
		},
		CustomData:   CustomData{
			UserId: user_id,
			Currency: "INR",
			Value: deposit_amount,
		},
		ActionSource: "app",
	}}

	cl := resty.New()
	_, err := cl.R().SetBody(pixelRequestBody).Post(os.Getenv("PIXEL_END_POINT"))
	if err != nil {
		log.Printf("Pixel FTD Api Call error for user %v for amount %v, %v", user_id, deposit_amount, err.Error())
	}
}