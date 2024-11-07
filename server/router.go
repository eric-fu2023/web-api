package server

import (
	"os"
	"time"

	"web-api/api"
	analyst_api "web-api/api/analyst"
	dc_api "web-api/api/dc"
	dollar_jackpot_api "web-api/api/dollar_jackpot"
	fb_api "web-api/api/fb"
	game_integration_api "web-api/api/game_integration"
	imsb_api "web-api/api/imsb"
	"web-api/api/mock"
	prediction_api "web-api/api/prediction"
	promotion_api "web-api/api/promotion"
	referral_alliance_api "web-api/api/referral_alliance"
	saba_api "web-api/api/saba"
	stream_game_api "web-api/api/stream_game"
	taya_api "web-api/api/taya"
	teamup_api "web-api/api/teamup"

	"web-api/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middleware.CorrelationID())
	r.Use(middleware.ErrorLogStatus())

	// initialize api for 3rd party integrations
	InitSdkApi(r)
	// initialize api for internal usage
	InitInternalApi(r)

	// middlewares order can't be changed
	r.Use(middleware.Cors())
	r.GET("/ts", api.Ts)
	// geolocations
	r.GET("/v1/geolocation", api.GeolocationGet)

	// beta
	// r.GET("/v1/home_banners", api.GetHomeBanners)
	// payment
	r.GET("/finpay_redirect", api.FinpayRedirect)
	r.POST("/finpay_redirect", api.FinpayRedirect)
	captcha := r.Group("/captcha")
	{
		captcha.POST("/get", api.CaptchaGet)
		captcha.POST("/check", api.CaptchaCheck)
	}

	// returns the random domains for API, logging and nami for mobile app, no encryption
	r.GET("/init_app", api.DomainInitApp)
	// returns the domain to redirect for web, no encryption
	r.GET("/route", api.DomainInitRoute)

	// all APIs below needs signature in the HTTP header
	r.Use(middleware.CheckSignature())
	// all APIs below will be encrypted
	r.Use(middleware.EncryptPayload())
	r.Use(middleware.Ip())
	r.Use(middleware.BrandAgent())
	r.Use(middleware.Timezone())
	r.Use(middleware.Location())
	r.Use(middleware.Locale())
	r.Use(middleware.AB())

	r.GET("/ping", api.Ping)
	// returns the domain to redirect + the random domains for API, logging and nami for web, with encryption
	r.GET("/init_web", api.DomainInitWeb)

	v1 := r.Group("/v1")
	{
		// no cache routes
		v1.POST("/sms_otp", api.SmsOtp)
		v1.POST("/email_otp", api.EmailOtp)
		v1.POST("/whatsapp_otp", api.WhatsAppOtp)
		v1.POST("/login_otp", api.UserLoginOtp)
		v1.POST("/login_password", api.UserLoginPassword)
		v1.POST("/password", middleware.CheckAuth(), api.UserSetPassword)
		v1.GET("/otp-check", api.VerifyOtp)

		v1.GET("/check-order", api.CheckOrder)
		v1.GET("/home_banners", api.GetHomeBanners)

		requireMobile := os.Getenv("REGISTER_REQUIRES_MOBILE") == "TRUE"
		bypassSetMobileOtpVerify := os.Getenv("REGISTER_NO_VERIFY_MOBILE_OTP") == "TRUE"
		v1.POST("/register", api.UserRegister(requireMobile, bypassSetMobileOtpVerify))
		v1.GET("/referral", middleware.Cache(10*time.Second, false), api.VerifyReferralCode)

		v1.GET("/config", api.Config)
		v1.GET("/app_update", middleware.Channel(), middleware.Cache(1*time.Minute, false), api.AppUpdate)
		v1.GET("/announcements", middleware.CheckAuth(), middleware.CacheForGuest(1*time.Minute), api.Announcements)
		v1.GET("/categories", middleware.Cache(1*time.Minute, false), api.CategoryList)
		v1.GET("/vendors", middleware.Cache(1*time.Minute, false), api.VendorList)
		v1.GET("/streams", middleware.Cache(1*time.Minute, true), api.StreamList)
		// this one can not have cache, because notification will need instant refresh
		v1.GET("/stream-announcements", api.StreamAnnouncementList)
		// v1.GET("/streamer", middleware.Cache(1*time.Minute, false), api.Streamer)
		v1.GET("/streamer", api.Streamer) // remove cache due to has_jackpot needs real time updates
		v1.GET("/topup-methods", middleware.CheckAuth(), api.TopupMethodList)
		v1.GET("/withdraw-methods", middleware.CheckAuth(), api.WithdrawMethodList)
		v1.GET("/avatars", middleware.Cache(1*time.Minute, false), api.AvatarList)
		v1.POST("/share", api.ShareCreate)
		v1.GET("/share", api.ShareGet)
		v1.GET("/games", middleware.Cache(1*time.Minute, false), api.GameList)
		v1.GET("/streamer_external_game", api.GameByStreamer)
		v1.GET("/room_chat/history", api.RoomChatHistory)
		v1.GET("/stream_game", stream_game_api.StreamGame)
		v1.GET("/stream_games", middleware.Cache(1*time.Minute, false), stream_game_api.StreamGameList)
		v1.GET("/game_categories", middleware.Cache(5*time.Minute, false), game_integration_api.GameCategoryList)
		v1.GET("/sub_games", middleware.Cache(5*time.Minute, false), game_integration_api.SubGames)
		v1.GET("/featured_games", middleware.Cache(5*time.Minute, false), game_integration_api.FeaturedGames)

		// v1.GET("/promotion/list", middleware.CheckAuth(), middleware.Cache(5*time.Second, false), promotion_api.GetCoverList)
		v1.GET("/promotion/list", middleware.CheckAuth(), promotion_api.GetCoverList)
		// v1.GET("/promotion/details", middleware.CheckAuth(), middleware.CacheForGuest(5*time.Minute), promotion_api.GetDetail)
		v1.GET("/promotion/details", middleware.CheckAuth(), promotion_api.GetDetail)
		v1.GET("/promotion/custom-details", middleware.CheckAuth(), middleware.RequestLogger("get custom promotion details"), promotion_api.GetCustomDetail)
		v1.GET("/promotion/categories", middleware.CheckAuth(), middleware.Cache(5*time.Minute, false), promotion_api.GetCategoryList)

		v1.GET("/rtc_token", middleware.CheckAuth(), api.RtcToken)
		v1.GET("/rtc_tokens", middleware.CheckAuth(), api.RtcTokens)

		v1.GET("/vips", middleware.Cache(5*time.Minute, false), api.VipLoad)
		popup := v1.Group("/popup")
		{
			popup.GET("/show", middleware.CheckAuth(), api.Show)
			popup.GET("/spin", middleware.CheckAuth(), api.Spin)
			popup.GET("/spin_result", middleware.AuthRequired(true, true), api.SpinResult)
			popup.GET("/spin_history", middleware.AuthRequired(true, true), api.SpinHistory)
		}

		pm := v1.Group("/pm")
		{
			pm.GET("/cs/history", middleware.CheckAuth(), api.CsHistory)
			pm.POST("/cs/send", middleware.CheckAuth(), api.CsSend)
		}

		saba := v1.Group("/saba")
		{
			saba.GET("/get_url", middleware.CheckAuth(), saba_api.GetUrl)
		}

		dc := v1.Group("/dc")
		{
			dc.GET("/fun_play", middleware.CheckAuth(), dc_api.FunPlay)
		}

		dj := v1.Group("/dollar_jackpot")
		{
			dj.GET("", middleware.CheckAuth(), dollar_jackpot_api.DollarJackpotGet) // can't have cache due to "total" value; cache at the query
			dj.GET("/winners", dollar_jackpot_api.DollarJackpotWinners)
			dj.GET("/history", middleware.AuthRequired(true, true), dollar_jackpot_api.DollarJackpotBetReport)
		}

		referralAlliance := v1.Group("/referral/alliance")
		{
			referralAlliance.GET("rankings", middleware.CheckAuth(), referral_alliance_api.GetRankings)
		}

		v1.GET("/gifts", middleware.Cache(1*time.Minute, false), api.GiftList)

		auth := v1.Group("/user")
		{
			userWithoutBrand := auth.Group("")
			userWithoutBrand.Use(middleware.AuthRequired(true, false))
			{
				userWithoutBrand.GET("/me", api.Me)
				userWithoutBrand.DELETE("/me", api.UserDelete)
				userWithoutBrand.DELETE("/logout", api.UserLogout)
				userWithoutBrand.POST("/finish_setup", api.UserFinishSetup)
				userWithoutBrand.GET("/check_username", api.UserCheckUsername)
				userWithoutBrand.POST("/check_password", api.UserCheckPassword)
				userWithoutBrand.GET("/silenced", api.Silenced)
				userWithoutBrand.GET("/followings", api.UserFollowingList)
			}
			user := auth.Group("")
			user.Use(middleware.AuthRequired(true, true))
			{
				user.POST("/profile", api.ProfileUpdate)
				user.POST("/clear_wager", api.ClearWager)
				user.POST("/nickname", api.NicknameUpdate)
				user.POST("/profile_pic", api.ProfilePicUpload)
				user.GET("/notifications", api.UserNotificationList)
				user.PUT("/notification/mark_read", api.UserNotificationMarkRead)
				user.GET("/counters", api.UserCounters)
				user.PUT("/fcm_token", api.FcmTokenUpdate)
				user.GET("/wallets", api.UserWallets)
				user.PUT("/sync_wallet", api.UserSyncWallet)
				user.PUT("/recall", api.UserRecallFund)

				user.GET("/following_ids", api.UserFollowingIdList)
				user.POST("/follow", api.UserFollowingAdd)
				user.DELETE("/follow", api.UserFollowingRemove)

				user.GET("/favourites", api.UserFavouriteList)
				user.POST("/favourite", api.UserFavouriteAdd)
				user.DELETE("/favourite", api.UserFavouriteRemove)

				user.PUT("/secondary-password", api.UserSetSecondaryPassword)
				user.PUT("/mobile", api.UserSetMobile)
				user.PUT("/email", api.UserSetEmail)

				user.GET("/orders", api.OrderList)
				user.GET("/recent_games", api.UserRecentGameList)

				user.POST("/withdraw-accounts", api.WthdrawAccountsAdd)
				user.DELETE("/withdraw-accounts", api.WthdrawAccountsRemove)
				user.GET("/withdraw-accounts", api.WthdrawAccountsList)

				user.GET("/otp-check", api.VerifyOtp)

				user.POST("/feedback", api.FeedbackAdd)

				user.GET("/vip-status", api.VipGet)
				user.GET("/vip-rebate-details", middleware.Cache(5*time.Minute, false), api.VipLoadRebateRule)
				user.GET("/vip-referral-alliance-reward-details", middleware.Cache(5*time.Minute, false), api.VipLoadReferralAllianceRewardRule)

				user.POST("/gift-send", middleware.CheckAuth(), api.GiftSend)
				user.GET("/gift-records", middleware.CheckAuth(), api.GiftRecordList)

				user.GET("/user-heartbeat", api.UserHeartbeat)

				taya := user.Group("/taya")
				{
					taya.GET("/token", taya_api.GetToken)
				}

				fb := user.Group("/fb")
				{
					fb.GET("/token", fb_api.GetToken)
				}

				dcUserGroup := user.Group("/dc")
				{
					dcUserGroup.GET("/get_url", dc_api.GetUrl)
				}

				imsb := user.Group("/imsb")
				{
					imsb.GET("/token", imsb_api.GetToken)
					imsb.POST("/apply_voucher", imsb_api.ApplyVoucher)
				}

				djUserGroup := user.Group("/dollar_jackpot")
				{
					djUserGroup.POST("/place_order", dollar_jackpot_api.PlaceOrder)
				}

				sg := user.Group("/stream_game")
				{
					sg.POST("/place_order", stream_game_api.PlaceOrder)
				}

				integration := user.Group("/game")
				{
					integration.GET("/url", game_integration_api.GetUrl)
				}

				kyc := user.Group("/kyc")
				{
					kyc.GET("", api.GetKyc)
					kyc.POST("", api.SubmitKyc)
				}

				cash := user.Group("/cash")
				{
					cash.POST("/top-up-orders", api.TopUpOrder)
					cash.POST("/withdraw-orders", api.WithdrawOrder)
					cash.GET("/orders", api.ListCashOrder)
				}

				promotion := user.Group("/promotion")
				{
					promotion.GET("/list", middleware.Cache(1*time.Minute, false), promotion_api.GetCoverList)
					promotion.GET("/details", middleware.RequestLogger("get promotion details"), promotion_api.GetDetail)
					promotion.POST("/claim", middleware.RequestLogger("promotion claim"), promotion_api.PromotionClaim)
					promotion.POST("/join", middleware.RequestLogger("promotion join"), promotion_api.PromotionJoin)
				}

				voucher := user.Group("/voucher")
				{
					// voucher.POST("/claim")
					voucher.GET("/list", promotion_api.VoucherList)
					// voucher.GET("/details")
					voucher.POST("/applicables", promotion_api.VoucherApplicable) // may not do
					voucher.POST("/pre-binding", promotion_api.VoucherPreBinding) // fb
					voucher.POST("/post-binding", mock.MockOK)                    //
				}

				achievement := user.Group("/achievement")
				{
					achievement.GET("/list", api.AchievementList)
					achievement.POST("/complete", api.AchievementComplete)
				}

				referralAllianceUserGroup := user.Group("/referral/alliance")
				{
					referralAllianceUserGroup.GET("/summary", referral_alliance_api.GetSummary)
					referralAllianceUserGroup.GET("/referrals", referral_alliance_api.ListReferrals)
					referralAllianceUserGroup.GET("/referral_summary", referral_alliance_api.GetReferralSummary)
					referralAllianceUserGroup.GET("/referral_reward_records", referral_alliance_api.GetReferralRewardRecords)
				}

				v2 := user.Group("/v2")
				{
					v2.GET("/notifications", api.UserNotificationList)
				}
			}
		}

		prediction := v1.Group("/prediction", middleware.CheckAuth())
		{
			prediction.GET("", prediction_api.GetPredictionDetail)
			prediction.GET("list", prediction_api.ListPredictions)
			prediction.POST("add", prediction_api.AddUserPrediction)
		}

		analyst := v1.Group("/analyst", middleware.CheckAuth())
		{
			analyst.GET("", analyst_api.GetAnalystDetail)
			analyst.GET("/list", middleware.Cache(1*time.Minute, false), analyst_api.ListAnalysts)
			analyst.GET("/following", middleware.AuthRequired(true, true), analyst_api.ListFollowingAnalysts)
			analyst.POST("/following", middleware.AuthRequired(true, true), analyst_api.ToggleFollowAnalyst)
			analyst.GET("/following-ids", middleware.AuthRequired(true, true), analyst_api.GetFollowingAnalystIdsList)
			analyst.GET("/achievement", analyst_api.GetAnalystAchievement)
		}

		teamup := v1.Group("/teamup", middleware.CheckAuth())
		{
			teamup.GET("/", middleware.AuthRequired(true, false), teamup_api.GetTeamUpItem)
			teamup.GET("/detail", teamup_api.GetTeamUpItem)
			teamup.GET("/start", middleware.AuthRequired(true, false), teamup_api.StartTeamUp)
			teamup.GET("/list", middleware.AuthRequired(true, false), teamup_api.ListAllTeamUp)
			teamup.GET("/others", teamup_api.OtherTeamups)
			teamup.GET("/contribute/list", teamup_api.ContributedList)
			teamup.POST("/slash", middleware.AuthRequired(true, false), teamup_api.SlashBet)

			teamup.GET("/spin", middleware.AuthRequired(true, false), teamup_api.TeamupSpin)
			teamup.GET("/spin/result", middleware.AuthRequired(true, false), teamup_api.TeamupSpinResult)

			// DEPRECATED
			// teamup.POST("/testdeposit", teamup_api.TestDeposit)
		}

		v1.GET("/user/heartbeat", middleware.AuthRequired(false, false), api.Heartbeat)
	}

	return r
}
