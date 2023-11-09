package util

import (
	"blgit.rfdev.tech/taya/game-service/dc"
	"blgit.rfdev.tech/taya/game-service/fb"
	"blgit.rfdev.tech/taya/game-service/saba"
	"os"
)

var FBFactory fb.FB
var SabaFactory saba.Saba
var DCFactory dc.Dc

func InitFbFactory() {
	FBFactory = fb.FB{
		BaseUrl:           os.Getenv("GAME_FB_BASE_URL"),
		MerchantId:        os.Getenv("GAME_FB_MERCHANT_ID"),
		MerchantApiSecret: os.Getenv("GAME_FB_MERCHANT_API_SECRET"),
		IsSandbox:         true,
	}
}

func InitSabaFactory() {
	SabaFactory = saba.Saba{
		BaseUrl:    os.Getenv("GAME_SABA_BASE_URL"),
		VendorId:   os.Getenv("GAME_SABA_VENDOR_ID"),
		OperatorId: os.Getenv("GAME_SABA_OPERATOR_ID"),
		IsSandbox:  true,
	}
}

func InitDcFactory() {
	DCFactory = dc.Dc{
		BaseUrl:   os.Getenv("GAME_DC_BASE_URL"),
		BrandId:   os.Getenv("GAME_DC_BRAND_ID"),
		ApiKey:    os.Getenv("GAME_DC_API_KEY"),
		IsSandbox: true,
	}
}
