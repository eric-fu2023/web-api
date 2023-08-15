package util

import "blgit.rfdev.tech/taya/game-service/fb"

var FBFactory fb.FB

func InitFbFactory() {
	FBFactory = fb.FB{
		MerchantId:        "1552945083054354433",
		MerchantApiSecret: "Lc63hMKwQz0R8Y4MbB7F6mhCbzLuZoU9",
		IsSandbox:         true,
	}
}
