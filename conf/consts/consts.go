package consts

const (
	KycStatusPending   = 1
	KycStatusCompleted = 2
	KycStatusRejected  = 3
	GinErrorKey        = "gin_error"
)

var (
	OrderTypeMap = map[int64]string{
		-1: OrderTypeWithdraw,
		1:  OrderTypeTopup,
		2:  OrderTypeTopup,
		3:  OrderTypeTopup,
		4:  OrderTypeTopup,
	}
	OrderTypeDetailMap = map[int64]string{
		-1: OrderTypeWithdraw,
		1:  OrderTypeTopup,
		2:  OrderTypeDepositBonus,
		3:  OrderTypeBetInsurance,
		4:  OrderTypeBeginnerBonus,
	}
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
	PlatformIdToDcPlatformId = map[int64]string{
		1: "pc",
		2: "mobile",
		3: "mobile",
		4: "mobile",
	}
	PlatformIdToGameVendorColumn = map[int64]string{
		1: "web",
		2: "h5",
		3: "android",
		4: "ios",
	}
	GameVendor = map[string]int64{
		"fb":             1,
		"saba":           2,
		"dc":             3,
		"taya":           4,
		"imsb":           5,
		"dollar_jackpot": 6,
		"stream_game":    7,
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
	UserRole = map[string]int64{
		"user":      1,
		"test_user": 2,
		"streamer":  3,
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
		"backend":  3,
		"username": 4,
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
	AnnouncementType = map[string]int64{
		"text":         1,
		"image":        2,
		"audioOverlay": 3,
		"download":     4,
		"gameLobby":    5,
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
	CountryISO = map[string]string{
		"1":   "PH",
		"2":   "AF",
		"3":   "AX",
		"4":   "AL",
		"5":   "DZ",
		"6":   "US",
		"7":   "AD",
		"8":   "AO",
		"9":   "AI",
		"10":  "AQ",
		"11":  "AG",
		"12":  "AR",
		"13":  "AM",
		"14":  "AW",
		"15":  "AU",
		"16":  "AT",
		"17":  "AZ",
		"18":  "BS",
		"19":  "BH",
		"20":  "BD",
		"21":  "BB",
		"22":  "BY",
		"23":  "BE",
		"24":  "BZ",
		"25":  "BJ",
		"26":  "BM",
		"27":  "BT",
		"28":  "BO",
		"29":  "BQ",
		"30":  "BA",
		"31":  "BW",
		"32":  "BV",
		"33":  "BR",
		"34":  "IO",
		"35":  "VG",
		"36":  "BN",
		"37":  "BG",
		"38":  "BF",
		"39":  "MM",
		"40":  "BI",
		"41":  "CV",
		"42":  "KH",
		"43":  "CM",
		"44":  "CA",
		"45":  "BQ",
		"46":  "KY",
		"47":  "CF",
		"48":  "TD",
		"49":  "CL",
		"50":  "CN",
		"51":  "TW",
		"52":  "CX",
		"53":  "CC",
		"54":  "CO",
		"55":  "KM",
		"56":  "CG",
		"57":  "CK",
		"58":  "CR",
		"59":  "CI",
		"60":  "HR",
		"61":  "CU",
		"62":  "CW",
		"63":  "CY",
		"64":  "CZ",
		"65":  "CD",
		"66":  "DK",
		"67":  "DJ",
		"68":  "DM",
		"69":  "DO",
		"70":  "TL",
		"71":  "EC",
		"72":  "EG",
		"73":  "SV",
		"74":  "GQ",
		"75":  "ER",
		"76":  "EE",
		"77":  "SZ",
		"78":  "ET",
		"79":  "FK",
		"80":  "FO",
		"81":  "FJ",
		"82":  "FI",
		"83":  "FR",
		"84":  "GF",
		"85":  "PF",
		"86":  "TF",
		"87":  "GA",
		"88":  "GM",
		"89":  "GE",
		"90":  "DE",
		"91":  "GH",
		"92":  "GI",
		"93":  "GB",
		"94":  "GR",
		"95":  "GL",
		"96":  "GD",
		"97":  "GP",
		"98":  "GU",
		"99":  "GT",
		"100": "GG",
		"101": "GN",
		"102": "GW",
		"103": "GY",
		"104": "HT",
		"105": "HM",
		"106": "VA",
		"107": "HN",
		"108": "HK",
		"109": "HU",
		"110": "IS",
		"111": "IN",
		"112": "ID",
		"113": "IR",
		"114": "IQ",
		"115": "IE",
		"116": "IM",
		"117": "IL",
		"118": "IT",
		"119": "CI",
		"120": "JM",
		"121": "SJ",
		"122": "JP",
		"123": "JE",
		"124": "JO",
		"125": "KZ",
		"126": "KE",
		"127": "KI",
		"128": "KP",
		"129": "KR",
		"130": "KW",
		"131": "KG",
		"132": "LA",
		"133": "LA",
		"134": "LV",
		"135": "LB",
		"136": "LS",
		"137": "LR",
		"138": "LY",
		"139": "LI",
		"140": "LT",
		"141": "LU",
		"142": "MO",
		"143": "MK",
		"144": "MG",
		"145": "MW",
		"146": "MY",
		"147": "MV",
		"148": "ML",
		"149": "MT",
		"150": "MH",
		"151": "MQ",
		"152": "MR",
		"153": "MU",
		"154": "YT",
		"155": "MX",
		"156": "FM",
		"157": "MD",
		"158": "MC",
		"159": "MN",
		"160": "ME",
		"161": "MS",
		"162": "MA",
		"163": "MZ",
		"164": "MM",
		"165": "NA",
		"166": "NR",
		"167": "NP",
		"168": "NL",
		"169": "NC",
		"170": "NZ",
		"171": "NI",
		"172": "NE",
		"173": "NG",
		"174": "NU",
		"175": "NF",
		"176": "MP",
		"177": "NO",
		"178": "OM",
		"179": "PK",
		"180": "PW",
		"181": "PS",
		"182": "PA",
		"183": "PG",
		"184": "PY",
		"185": "PE",
		"186": "PN",
		"187": "PL",
		"188": "PT",
		"189": "PR",
		"190": "QA",
		"191": "RE",
		"192": "RO",
		"193": "RU",
		"194": "RW",
		"195": "BQ",
		"196": "EH",
		"197": "BL",
		"198": "SH",
		"199": "AC",
		"200": "TA",
		"201": "KN",
		"202": "LC",
		"203": "MF",
		"204": "PM",
		"205": "VC",
		"206": "WS",
		"207": "SM",
		"208": "ST",
		"209": "SA",
		"210": "SN",
		"211": "RS",
		"212": "SC",
		"213": "SL",
		"214": "SG",
		"215": "BQ",
		"216": "SX",
		"217": "SK",
		"218": "SI",
		"219": "SB",
		"220": "SO",
		"221": "ZA",
		"222": "GS",
		"223": "SS",
		"224": "ES",
		"225": "LK",
		"226": "SD",
		"227": "SR",
		"228": "SJ",
		"229": "SE",
		"230": "CH",
		"231": "SY",
		"232": "TJ",
		"233": "TZ",
		"234": "TH",
		"235": "TL",
		"236": "TG",
		"237": "TK",
		"238": "TO",
		"239": "TT",
		"240": "TN",
		"241": "TR",
		"242": "TM",
		"243": "TC",
		"244": "TV",
		"245": "UG",
		"246": "UA",
		"247": "UY",
		"248": "UZ",
		"249": "VU",
		"250": "VA",
		"251": "VE",
		"252": "VN",
		"253": "VI",
		"254": "WF",
		"255": "EH",
		"256": "YE",
		"257": "ZM",
		"258": "ZW",
	}
)

const (
	CorrelationHeader             = "X-Correlation-ID"
	LogKey                        = "logger"
	CorrelationKey                = "correlation_id"
	StdTimeFormat                 = "2006-01-02 15:04:05"
	OrderTypeTopup                = "top-up"
	OrderTypeWithdraw             = "withdraw"
	OrderTypeDepositBonus         = "deposit_bonus"
	OrderTypeBetInsurance         = "bet_insurance"
	OrderTypeBeginnerBonus        = "beginner_reward"
	ConfigKeyTopupKycCheck string = "kyc_check_required"
	// FirstTopupMinimum      int64  = 10_00 //1000_00
	// TopupMinimum           int64  = 5_00  //500_00
	// TopupMax               int64  = 30000_00
	WithdrawMethodLimit = 5
	StackTraceKey       = "stack_trace"

	Notification_Type_User_Registration = "user_registration"
	Notification_Type_Password_Reset    = "password_reset"
	Notification_Type_Pin_Reset         = "pin_reset"
	Notification_Type_Kyc               = "kyc"
	Notification_Type_Bet_Placement     = "bet_transaction"
	Notification_Type_Bet_Settlement    = "bet_settlement"
	Notification_Type_Cash_Transaction  = "cash_transaction"
	Notification_Type_Deposit_Bonus     = "deposit_bonus"
	Notification_Type_Mobile_Reset      = "mobile_reset"

	//DefaultBrand = 1001
	//DefaultAgent = 1000001

	SmsOtpActionLogin                = "login"
	SmsOtpActionDeleteUser           = "delete_user"
	SmsOtpActionSetPassword          = "set_password"
	SmsOtpActionSetSecondaryPassword = "set_secondary_password"
	SmsOtpActionSetMobile            = "set_mobile"
)
