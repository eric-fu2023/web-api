package service

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"
	"web-api/serializer"
	"web-api/util"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)
type PixelInstall struct {
}
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

func (s PixelInstall)PixelInstallEvent(c *gin.Context)(r serializer.Response, err error){
	device_info,err:= util.GetDeviceInfo(c)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("sending pixel app data for install event, get device info err: %s", err.Error())
		r = serializer.GeneralErr(c, err)
		return
	}
	if device_info.Channel == "pixel_app_001"{
		log.Printf("should log pixel event install for channel pixel_app_001")
		PixelInstallEvent(c.ClientIP(), os.Getenv("PIXEL_ACCESS_TOKEN"), os.Getenv("PIXEL_END_POINT"))
	}
	if device_info.Channel == "pixel_app_002"{
		log.Printf("should log pixel event install for channel pixel_app_002")
		PixelInstallEvent(c.ClientIP(), os.Getenv("PIXEL_ACCESS_TOKEN_002"), os.Getenv("PIXEL_END_POINT_002"))
	}
	return 
}

func PixelInstallEvent(client_ip string, token string, url string) {
	var pixelRequestBody PixelRequestBody
	pixelRequestBody.AccessToken = token
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
	resp, err := cl.R().SetBody(pixelRequestBody).Post(url)
	log.Printf("pixel resp: %v", resp.String())
	if err != nil {
		log.Printf("Pixel Install Api Call error for user %v", err.Error())
	}
}

func PixelRegisterEvent(user_id int64, client_ip string, token string, post_url string, pixel_id string) {
	var pixelRequestBody PixelRequestBody
	pixelRequestBody.AccessToken = token
	pixelRequestBody.Data = []PixelRequestBodyData{{
		EventName:    "CompleteRegistration",
		EventTime:    time.Now().Unix(),
		UserData:     UserData{
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
	resp, err := cl.R().SetBody(pixelRequestBody).Post(post_url)
	if err != nil {
		log.Printf("Pixel Register Api Call error for user %v, %v", user_id, err.Error())
	} else {
		log.Printf("pixel resp: %v", resp.String())
	}


	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Define the number of digits
	digitCount := 17

	// Generate a random 17-digit string
	var sb strings.Builder
	for i := 0; i < digitCount; i++ {
		if i == 0 {
			// Ensure the first digit is not zero
			sb.WriteByte(byte(rand.Intn(9) + 1 + '0'))
		} else {
			// Subsequent digits can include zero
			sb.WriteByte(byte(rand.Intn(10) + '0'))
		}
	}
	randomString := sb.String()

	// Encode query parameters
	query := url.Values{}
	query.Add("id", pixel_id)
	query.Add("ev", "CompleteRegistration")
	query.Add("dl", "https://googleplay.batacezoom.com/")
	query.Add("if", "false")
	query.Add("ec", "0")
	query.Add("ts", time.Now().Format("20060102150405")) // Example timestamp
	query.Add("sw", "375")
	query.Add("sh", "667")
	query.Add("v", "2.9.176")
	query.Add("o", "4126")
	query.Add("ler", "empty")
	query.Add("coo", "false")
	query.Add("cdl", "")
	query.Add("rl", "")
	query.Add("fbp", fmt.Sprintf("fb.1.%s.%s", time.Now().Format("20060102150405"), randomString))
	query.Add("rqm", "GET")
	get_url := "https://www.facebook.com/privacy_sandbox/pixel/register/trigger/?" + query.Encode()

	// Make the GET request
	get_cl := resty.New()
	get_resp, get_err := get_cl.R().Get(get_url)

	if get_err != nil {
		log.Printf("Pixel Register API Call error for user %v, %v", user_id, err.Error())
	} else {
		log.Printf("Pixel get_resp: %v", get_resp.String())
	}
}

func PixelFTDEvent(user_id int64, client_ip string, deposit_amount int64, token string, post_url string, pixel_id string) {
	var pixelRequestBody PixelRequestBody
	pixelRequestBody.AccessToken = token
	pixelRequestBody.Data = []PixelRequestBodyData{{
		EventName:    "Purchase",
		EventTime:    time.Now().Unix(),
		UserData:     UserData{
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
	resp, err := cl.R().SetBody(pixelRequestBody).Post(post_url)
	log.Printf("pixel resp: %v", resp.String())

	if err != nil {
		log.Printf("Pixel FTD Api Call error for user %v for amount %v, %v", user_id, deposit_amount, err.Error())
	}

		// Seed the random number generator
		rand.Seed(time.Now().UnixNano())

		// Define the number of digits
		digitCount := 17
	
		// Generate a random 17-digit string
		var sb strings.Builder
		for i := 0; i < digitCount; i++ {
			if i == 0 {
				// Ensure the first digit is not zero
				sb.WriteByte(byte(rand.Intn(9) + 1 + '0'))
			} else {
				// Subsequent digits can include zero
				sb.WriteByte(byte(rand.Intn(10) + '0'))
			}
		}
		randomString := sb.String()
	
		// Encode query parameters
		query := url.Values{}
		query.Add("id", pixel_id)
		query.Add("ev", "Purchase")
		query.Add("dl", "https://googleplay.batacezoom.com/")
		query.Add("if", "false")
		query.Add("ec", "0")
		query.Add("ts", time.Now().Format("20060102150405")) // Example timestamp
		query.Add("sw", "375")
		query.Add("sh", "667")
		query.Add("v", "2.9.176")
		query.Add("o", "4126")
		query.Add("ler", "empty")
		query.Add("coo", "false")
		query.Add("cdl", "")
		query.Add("rl", "")
		query.Add("fbp", fmt.Sprintf("fb.1.%s.%s", time.Now().Format("20060102150405"), randomString))
		query.Add("rqm", "GET")
		get_url := "https://www.facebook.com/privacy_sandbox/pixel/register/trigger/?" + query.Encode()
	
		// Make the GET request
		get_cl := resty.New()
		get_resp, get_err := get_cl.R().Get(get_url)
	
		if get_err != nil {
			log.Printf("Pixel Register API Call error for user %v, %v", user_id, err.Error())
		} else {
			log.Printf("Pixel get_resp: %v", get_resp.String())
		}
}