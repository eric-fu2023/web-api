package consts

const (
	KycStatusPending   = 1
	KycStatusCompleted = 2
	GinErrorKey        = "gin_error"
)

var (
	Platform = map[string]int64{
		"pc":      1,
		"h5":      2,
		"android": 3,
		"ios":     4,
	}
	PlatformIdToFbPlatformId = map[int64]string{
		1: "pc",
		2: "h5",
		3: "mobile",
		4: "mobile",
	}
	PlatformIdToSabaPlatformId = map[int64]string{
		1: "1",
		2: "2",
		3: "3",
		4: "3",
	}
	GameVendor = map[string]int64{
		"fb":   1,
		"saba": 2,
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
	ChatSystemId  int64 = -999999
	AuthEventType       = map[string]int{
		"login":  1,
		"logout": 2,
	}

	AuthEventStatus = map[string]int{
		"successful": 1,
		"failed":     2,
	}

	AuthEventLoginMethod = map[string]int{
		"otp":      1,
		"password": 2,
	}
	SportsType = map[string]int64{
		"football":   1,
		"basketball": 2,
		"tennis":     3,
		"replay":     4,
		"worldcup":   7,
		"nba":        8,
		"volleyball": 10,
		"lol":        101,
		"csgo":       102,
		"dota2":      103,
	}
	CashOrderStatus = map[int64]string{
		1: "Pending",
		2: "Completed",
		3: "Failed",
		4: "Pending", // approval
		5: "Failed",
		6: "Pending",
		7: "Failed",
	}
)

const (
	CorrelationHeader             = "X-Correlation-ID"
	LogKey                        = "logger"
	CorrelationKey                = "correlation_id"
	StdTimeFormat                 = "2006-01-02 15:04:05"
	OrderTypeTopup                = "top-up"
	OrderTypeWithdraw             = "withdraw"
	ConfigKeyTopupKycCheck string = "kyc_check_required"
	FirstTopupMinimum      int64  = 1000_00
	TopupMinimum           int64  = 500_00
)
