package consts

import (
	"web-api/util/datastructures"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

const (
	InHouseGame  = 1 // this is internal self-dev games such as dice
	ExternalGame = 2 // this is 3rd party sub games
)

const (
	KycStatusPending   = 1
	KycStatusCompleted = 2
	KycStatusRejected  = 3
	GinErrorKey        = "gin_error"
)

var (
	SpinResultType = map[int64]string{
		1: SpinVoucher,
		2: SpinBonus,
		3: SpinItem,
		4: SpinOthers,
	}
	OrderTypeMap = map[int64]string{
		-1: OrderTypeWithdraw,
		1:  OrderTypeTopup,
		2:  OrderTypeTopup,
		3:  OrderTypeTopup,
		4:  OrderTypeTopup,
		5:  OrderTypeTopup,
		6:  OrderTypeTopup,
		7:  OrderTypeTopup,
		8:  OrderTypeTopup,
		9:  OrderTypeTopup,
		10: OrderTypeTopup,
		11: OrderTypeTopup,
		12: OrderTypeTopup,
		13: OrderTypeTopup,
		14: OrderTypeTopup,
		15: OrderTypeTopup,
	}
	OrderTypeDetailMap = map[int64]string{
		-1: OrderTypeWithdraw,
		1:  OrderTypeTopup,
		2:  OrderTypeDepositBonus,
		3:  OrderTypeBetInsurance,
		4:  OrderTypeBeginnerBonus,
		5:  OrderTypeVipBday,
		6:  OrderTypeVipPromo,
		7:  OrderTypeVipWeekly,
		8:  OrderTypeVipRebate,
		9:  OrderTypeVipReferral,
		10: OrderTypeVipCashMethodPromotion,
		11: OrderTypeTopup,
		12: OrderTypeTeamupRebate,
		13: OrderTypeSpinRewards,
	}
	OrderOperationTypeDetailMap = map[models.OperationType]string{
		models.CashOrderOperationTypeSystemAdjust:       OrderOperationTypeSystemAdjust,
		models.CashOrderOperationTypeGameAdjust:         OrderOperationTypeGameAdjust,
		models.CashOrderOperationTypeCashInAdjust:       OrderOperationTypeCashInAdjust,
		models.CashOrderOperationTypeBonus:              OrderOperationTypeBonus,
		models.CashOrderOperationTypeVipUpgrade:         OrderOperationTypeVipUpgrade,
		models.CashOrderOperationTypeWeeklyBonus:        OrderOperationTypeWeeklyBonus,
		models.CashOrderOperationTypeBirthdayBonus:      OrderOperationTypeBirthdayBonus,
		models.CashOrderOperationTypeAgentBonus:         OrderOperationTypeAgentBonus,
		models.CashOrderOperationTypePromoteBonus:       OrderOperationTypePromoteBonus,
		models.CashOrderOperationTypeDepositBonus:       OrderOperationTypeDepositBonus,
		models.CashOrderOperationTypeEventBonus:         OrderOperationTypeEventBonus,
		models.CashOrderOperationTypeReset:              OrderOperationTypeReset,
		models.CashOrderOperationTypeReferralBonus:      OrderOperationTypeReferralBonus,
		models.CashOrderOperationTypeCashOutBonus:       OrderOperationTypeCashOutBonus,
		models.CashOrderOperationTypeMakeUpOrder:        OrderOperationTypeMakeUpOrder,
		models.CashOrderOperationTypeWithdrawalReversal: OrderOperationTypeWithdrawalReversal,
	}

	// OrderOperationTypeEnum tapes OrderOperationTypeDetailMap
	OrderOperationTypeEnum = datastructures.KVtoVK(OrderOperationTypeDetailMap)

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
		"join":       7,
		"typing":     100,
		"empty":      101,
		"gift":       10,
	}
	UserRole = map[string]int64{
		"user":      1,
		"test_user": 2,
		"streamer":  3,
	}
	ChatUserType = map[string]int64{
		"admin":     1,
		"streamer":  2,
		"user":      3,
		"guest":     4,
		"assistant": 5,
		"bot":       6,
	}
	ChatSystemId  int64 = -999999
	AuthEventType       = map[string]int{
		"login":         1,
		"logout":        2,
		"forced_logout": 3,
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
		8: "Pending", // pending risk check
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
	CountryMap = map[string]string{
		"BD": "Bangladesh",
		"BE": "Belgium",
		"BF": "Burkina Faso",
		"BG": "Bulgaria",
		"BA": "Bosnia and Herzegovina",
		"BB": "Barbados",
		"WF": "Wallis and Futuna",
		"BL": "Saint Barthelemy",
		"BM": "Bermuda",
		"BN": "Brunei",
		"BO": "Bolivia",
		"BH": "Bahrain",
		"BI": "Burundi",
		"BJ": "Benin",
		"BT": "Bhutan",
		"JM": "Jamaica",
		"BV": "Bouvet Island",
		"BW": "Botswana",
		"WS": "Samoa",
		"BQ": "Bonaire, Saint Eustatius and Saba ",
		"BR": "Brazil",
		"BS": "Bahamas",
		"JE": "Jersey",
		"BY": "Belarus",
		"BZ": "Belize",
		"RU": "Russia",
		"RW": "Rwanda",
		"RS": "Serbia",
		"TL": "East Timor",
		"RE": "Reunion",
		"TM": "Turkmenistan",
		"TJ": "Tajikistan",
		"RO": "Romania",
		"TK": "Tokelau",
		"GW": "Guinea-Bissau",
		"GU": "Guam",
		"GT": "Guatemala",
		"GS": "South Georgia and the South Sandwich Islands",
		"GR": "Greece",
		"GQ": "Equatorial Guinea",
		"GP": "Guadeloupe",
		"JP": "Japan",
		"GY": "Guyana",
		"GG": "Guernsey",
		"GF": "French Guiana",
		"GE": "Georgia",
		"GD": "Grenada",
		"GB": "United Kingdom",
		"GA": "Gabon",
		"SV": "El Salvador",
		"GN": "Guinea",
		"GM": "Gambia",
		"GL": "Greenland",
		"GI": "Gibraltar",
		"GH": "Ghana",
		"OM": "Oman",
		"TN": "Tunisia",
		"JO": "Jordan",
		"HR": "Croatia",
		"HT": "Haiti",
		"HU": "Hungary",
		"HK": "Hong Kong",
		"HN": "Honduras",
		"HM": "Heard Island and McDonald Islands",
		"VE": "Venezuela",
		"PR": "Puerto Rico",
		"PS": "Palestinian Territory",
		"PW": "Palau",
		"PT": "Portugal",
		"SJ": "Svalbard and Jan Mayen",
		"PY": "Paraguay",
		"IQ": "Iraq",
		"PA": "Panama",
		"PF": "French Polynesia",
		"PG": "Papua New Guinea",
		"PE": "Peru",
		"PK": "Pakistan",
		"PH": "Philippines",
		"PN": "Pitcairn",
		"PL": "Poland",
		"PM": "Saint Pierre and Miquelon",
		"ZM": "Zambia",
		"EH": "Western Sahara",
		"EE": "Estonia",
		"EG": "Egypt",
		"ZA": "South Africa",
		"EC": "Ecuador",
		"IT": "Italy",
		"VN": "Vietnam",
		"SB": "Solomon Islands",
		"ET": "Ethiopia",
		"SO": "Somalia",
		"ZW": "Zimbabwe",
		"SA": "Saudi Arabia",
		"ES": "Spain",
		"ER": "Eritrea",
		"ME": "Montenegro",
		"MD": "Moldova",
		"MG": "Madagascar",
		"MF": "Saint Martin",
		"MA": "Morocco",
		"MC": "Monaco",
		"UZ": "Uzbekistan",
		"MM": "Myanmar",
		"ML": "Mali",
		"MO": "Macao",
		"MN": "Mongolia",
		"MH": "Marshall Islands",
		"MK": "Macedonia",
		"MU": "Mauritius",
		"MT": "Malta",
		"MW": "Malawi",
		"MV": "Maldives",
		"MQ": "Martinique",
		"MP": "Northern Mariana Islands",
		"MS": "Montserrat",
		"MR": "Mauritania",
		"IM": "Isle of Man",
		"UG": "Uganda",
		"TZ": "Tanzania",
		"MY": "Malaysia",
		"MX": "Mexico",
		"IL": "Israel",
		"FR": "France",
		"IO": "British Indian Ocean Territory",
		"SH": "Saint Helena",
		"FI": "Finland",
		"FJ": "Fiji",
		"FK": "Falkland Islands",
		"FM": "Micronesia",
		"FO": "Faroe Islands",
		"NI": "Nicaragua",
		"NL": "Netherlands",
		"NO": "Norway",
		"NA": "Namibia",
		"VU": "Vanuatu",
		"NC": "New Caledonia",
		"NE": "Niger",
		"NF": "Norfolk Island",
		"NG": "Nigeria",
		"NZ": "New Zealand",
		"NP": "Nepal",
		"NR": "Nauru",
		"NU": "Niue",
		"CK": "Cook Islands",
		"XK": "Kosovo",
		"CI": "Ivory Coast",
		"CH": "Switzerland",
		"CO": "Colombia",
		"CN": "China",
		"CM": "Cameroon",
		"CL": "Chile",
		"CC": "Cocos Islands",
		"CA": "Canada",
		"CG": "Republic of the Congo",
		"CF": "Central African Republic",
		"CD": "Democratic Republic of the Congo",
		"CZ": "Czech Republic",
		"CY": "Cyprus",
		"CX": "Christmas Island",
		"CR": "Costa Rica",
		"CW": "Curacao",
		"CV": "Cape Verde",
		"CU": "Cuba",
		"SZ": "Swaziland",
		"SY": "Syria",
		"SX": "Sint Maarten",
		"KG": "Kyrgyzstan",
		"KE": "Kenya",
		"SS": "South Sudan",
		"SR": "Suriname",
		"KI": "Kiribati",
		"KH": "Cambodia",
		"KN": "Saint Kitts and Nevis",
		"KM": "Comoros",
		"ST": "Sao Tome and Principe",
		"SK": "Slovakia",
		"KR": "South Korea",
		"SI": "Slovenia",
		"KP": "North Korea",
		"KW": "Kuwait",
		"SN": "Senegal",
		"SM": "San Marino",
		"SL": "Sierra Leone",
		"SC": "Seychelles",
		"KZ": "Kazakhstan",
		"KY": "Cayman Islands",
		"SG": "Singapore",
		"SE": "Sweden",
		"SD": "Sudan",
		"DO": "Dominican Republic",
		"DM": "Dominica",
		"DJ": "Djibouti",
		"DK": "Denmark",
		"VG": "British Virgin Islands",
		"DE": "Germany",
		"YE": "Yemen",
		"DZ": "Algeria",
		"US": "United States",
		"UY": "Uruguay",
		"YT": "Mayotte",
		"UM": "United States Minor Outlying Islands",
		"LB": "Lebanon",
		"LC": "Saint Lucia",
		"LA": "Laos",
		"TV": "Tuvalu",
		"TW": "Taiwan",
		"TT": "Trinidad and Tobago",
		"TR": "Turkey",
		"LK": "Sri Lanka",
		"LI": "Liechtenstein",
		"LV": "Latvia",
		"TO": "Tonga",
		"LT": "Lithuania",
		"LU": "Luxembourg",
		"LR": "Liberia",
		"LS": "Lesotho",
		"TH": "Thailand",
		"TF": "French Southern Territories",
		"TG": "Togo",
		"TD": "Chad",
		"TC": "Turks and Caicos Islands",
		"LY": "Libya",
		"VA": "Vatican",
		"VC": "Saint Vincent and the Grenadines",
		"AE": "United Arab Emirates",
		"AD": "Andorra",
		"AG": "Antigua and Barbuda",
		"AF": "Afghanistan",
		"AI": "Anguilla",
		"VI": "U.S. Virgin Islands",
		"IS": "Iceland",
		"IR": "Iran",
		"AM": "Armenia",
		"AL": "Albania",
		"AO": "Angola",
		"AQ": "Antarctica",
		"AS": "American Samoa",
		"AR": "Argentina",
		"AU": "Australia",
		"AT": "Austria",
		"AW": "Aruba",
		"IN": "India",
		"AX": "Aland Islands",
		"AZ": "Azerbaijan",
		"IE": "Ireland",
		"ID": "Indonesia",
		"UA": "Ukraine",
		"QA": "Qatar",
		"MZ": "Mozambique",
	}
)

const (
	CorrelationHeader                           = "X-Correlation-ID"
	ClientIpHeader                              = "X-Forwarded-For"
	LogKey                                      = "logger"
	CorrelationKey                              = "correlation_id"
	StdTimeFormat                               = "2006-01-02 15:04:05"
	StdMonthFormat                              = "2006-01"
	OrderTypeTopup                              = "top-up"
	OrderTypeWithdraw                           = "withdraw"
	OrderTypeDepositBonus                       = "deposit_bonus"
	OrderTypeBetInsurance                       = "bet_insurance"
	OrderTypeBeginnerBonus                      = "beginner_reward"
	ConfigKeyTopupKycCheck               string = "kyc_check_required"
	OrderTypeVipBday                            = "vip_bday"
	OrderTypeVipPromo                           = "vip_promo"
	OrderTypeVipWeekly                          = "vip_weekly"
	OrderTypeVipRebate                          = "vip_rebate"
	OrderTypeVipReferral                        = "vip_referral"
	OrderTypeVipCashMethodPromotion             = "cash_method_promotion"
	OrderOperationTypeSystemAdjust              = "system_adjust"
	OrderOperationTypeGameAdjust                = "game_adjust"
	OrderOperationTypeCashInAdjust              = "cash_in_adjust"
	OrderOperationTypeBonus                     = "bonus"
	OrderOperationTypeVipUpgrade                = "vip_upgrade_bonus"
	OrderOperationTypeWeeklyBonus               = "weekly_bonus"
	OrderOperationTypeBirthdayBonus             = "birthday_bonus"
	OrderOperationTypeAgentBonus                = "agent_bonus"
	OrderOperationTypePromoteBonus              = "promote_bonus"
	OrderOperationTypeDepositBonus              = "deposit_bonus"
	OrderOperationTypeEventBonus                = "event_bonus"
	OrderOperationTypeReset                     = "reset"
	OrderOperationTypeReferralBonus             = "referral_bonus"
	OrderOperationTypeCashOutBonus              = "cash_out_bonus"
	OrderOperationTypeMakeUpOrder               = "make_up_order"
	OrderOperationTypeWithdrawalReversal        = "withdrawal_reversal"
	OrderTypeTeamupRebate                       = "teamup_rebate"
	OrderTypeSpinRewards                        = "spin_rewards"
	// FirstTopupMinimum      int64  = 10_00 //1000_00
	// TopupMinimum           int64  = 5_00  //500_00
	// TopupMax               int64  = 30000_00
	WithdrawMethodLimit = 5
	StackTraceKey       = "stack_trace"

	Notification_Type_User_Registration   = "user_registration"
	Notification_Type_Password_Reset      = "password_reset"
	Notification_Type_Pin_Reset           = "pin_reset"
	Notification_Type_Kyc                 = "kyc"
	Notification_Type_Bet_Placement       = "bet_transaction"
	Notification_Type_Bet_Settlement      = "bet_settlement"
	Notification_Type_Cash_Transaction    = "cash_transaction"
	Notification_Type_Deposit_Bonus       = "deposit_bonus"
	Notification_Type_Mobile_Reset        = "mobile_reset"
	Notification_Type_Email_Reset         = "email_reset"
	Notification_Type_Birthday_Bonus      = "birthday_bonus"
	Notification_Type_Vip_Promotion_Bonus = "vip_promotion_bonus"
	Notification_Type_Weekly_Bonus        = "weekly_bonus"
	Notification_Type_Rebate              = "rebate"
	Notification_Type_Vip_Promotion       = "vip_promotion"
	Notification_Type_Pop_Up              = "popup_winlose"
	Notification_Type_Spin                = "spin"
	Notification_Type_Referral_Alliance   = "referral_alliance"
	Notification_Type_Teamup              = "teamup"
	Notification_Type_Teamup_Detail       = "teamup_detail"

	//DefaultBrand = 1001
	//DefaultAgent = 1000001

	SmsOtpActionLogin                = "login"
	SmsOtpActionDeleteUser           = "delete_user"
	SmsOtpActionSetPassword          = "set_password"
	SmsOtpActionSetSecondaryPassword = "set_secondary_password"
	SmsOtpActionSetMobile            = "set_mobile"
	SmsOtpActionSetEmail             = "set_email"

	// spin related
	SpinVoucher = "spin_voucher"
	SpinBonus   = "spin_bonus"
	SpinItem    = "spin_item"
	SpinOthers  = "spin_others"
)

// 游戏砍单有关
// 游戏砍单名称 / 图标
var (
	GameProviderNameMap = map[string]string{
		"ba01inr":          "BatAce Original",
		"SPRIBE_SLOT":      "Spribe",
		"MICROGAMING_SLOT": "Microgaming",
		"HS_SLOT":          "Hacksaw Gaming",
		"JFI0000":          "JFI",
		"PLAYNGO_SLOT":     "Play'n GO",
		"evolution":        "Evolution",
		"Evolution":        "Evolution",
		"IMOne Slot":       "IMOne Slot",
		"Mumbai":           "Mumbai",
		"9Wickets":         "9Wickets",
	}

	GameProviderNameToImgMap = map[string]string{
		"ba01inr":          "",
		"SPRIBE_SLOT":      "",
		"MICROGAMING_SLOT": "",
		"HS_SLOT":          "",
		"JFI0000":          "",
		"PLAYNGO_SLOT":     "",
		"evolution":        "https://cdn.tayalive.com/temp/aha/stream/7.jpg",
		"Evolution":        "https://cdn.tayalive.com/temp/aha/stream/7.jpg",
		"IMOne Slot":       "https://cdn.tayalive.com/temp/aha/stream/7.jpg",
		"Mumbai":           "https://cdn.tayalive.com/temp/aha/stream/7.jpg",
		"9Wickets":         "https://cdn.tayalive.com/temp/aha/stream/7.jpg",
	}
)
