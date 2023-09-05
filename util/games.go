package util

import (
	"blgit.rfdev.tech/taya/game-service/fb"
	"blgit.rfdev.tech/taya/game-service/saba"
	"os"
)

var FBFactory fb.FB
var SabaFactory saba.Saba

func InitFbFactory() {
	FBFactory = fb.FB{
		MerchantId:        os.Getenv("GAME_FB_MERCHANT_ID"),
		MerchantApiSecret: os.Getenv("GAME_FB_MERCHANT_API_SECRET"),
		IsSandbox:         true,
	}
}

func InitSabaFactory() {
	SabaFactory = saba.Saba{
		VendorId:   os.Getenv("GAME_SABA_VENDOR_ID"),
		OperatorId: os.Getenv("GAME_SABA_OPERATOR_ID"),
		IsSandbox:  true,
	}
}
