package util

import (
	"blgit.rfdev.tech/taya/game-service/fb"
	"os"
)

var FBFactory fb.FB

func InitFbFactory() {
	FBFactory = fb.FB{
		MerchantId:        os.Getenv("GAME_FB_MERCHANT_ID"),
		MerchantApiSecret: os.Getenv("GAME_FB_MERCHANT_API_SECRET"),
		IsSandbox:         true,
	}
}
