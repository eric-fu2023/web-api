package consts

var (
	Platform = map[int64]string{
		1: "pc",
		2: "h5",
		3: "android",
		4: "ios",
	}
	PlatformIdToFbPlatformId = map[int64]string{
		1: "pc",
		2: "h5",
		3: "mobile",
		4: "mobile",
	}
	GameProvider = map[string]int64{
		"fb": 1,
	}
	WebSocketMessageType = map[string]int64{
		"text":       1,
		"pic_url":    2,
		"pic_base64": 3,
		"ad_text":    4,
		"sticker":    5,
		"typing":     100,
		"empty":      101,
	}
	ChatUserType = map[string]int64{
		"system":    1,
		"bot":       2,
		"user":      3,
		"guest":     4,
		"assistant": 5,
		"streamer":  6,
	}
	ChatSystem = map[string][]string{
		"names": {"System"},
		"messages": {
			"Welcome to the chat room, where everyone can talk about sports. The platform administrator conducts 24-hour online inspections. If there are any violations of laws and regulations, pornography and vulgarity in the chat room, please report it immediately. The messages in the chat room only represent personal opinions and do not represent the opinions of the platform. Do not give your account to other people.",
			"Connecting...",
			"Successfully connected.",
		},
	}
	ChatSystemId int64 = -999999
	FbTransferTypeCalculateWager = map[string]int64{
		"WIN": -1,
		"CASHOUT": -1,
		"SETTLEMENT_ROLLBACK_RETURN": 1,
		"SETTLEMENT_ROLLBACK_DEDUCT": 1,
		"CASHOUT_CANCEL_DEDUCT": 1,
		"CASHOUT_CANCEL_RETURN": 1,
		"CASHOUT_CANCEL_ROLLBACK_DEDUCT": -1,
		"CASHOUT_CANCEL_ROLLBACK_RETURN": -1,
	}
)
