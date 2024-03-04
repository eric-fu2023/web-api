package server

import (
	"os"
	"time"
	"web-api/api"
	dc_api "web-api/api/dc"
	"web-api/api/dollar_jackpot"
	fb_api "web-api/api/fb"
	api_finpay "web-api/api/finpay"
	imsb_api "web-api/api/imsb"
	internal_api "web-api/api/internalapi"
	"web-api/api/mock"
	promotion_api "web-api/api/promotion"
	saba_api "web-api/api/saba"
	taya_api "web-api/api/taya"

	"web-api/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.New()

	r.Use(middleware.CorrelationID())
	r.Use(middleware.ErrorLogStatus())

	if os.Getenv("GAME_TAYA_EXPOSE_CALLBACKS") == "true" {
		fbCallback := r.Group("/taya/fb/callback")
		{
			fbCallback.POST("/health", taya_api.CallbackHealth)
			fbCallback.POST("/balance", taya_api.CallbackBalance)
			fbCallback.POST("/order_pay", taya_api.CallbackOrderPay)
			fbCallback.POST("/check_order_pay", taya_api.CallbackCheckOrderPay)
			fbCallback.POST("/sync_transaction", taya_api.CallbackSyncTransaction)
			fbCallback.POST("/sync_orders", taya_api.CallbackSyncOrders)
			fbCallback.POST("/sync_cashout", taya_api.CallbackSyncCashout)
		}
	}

	if os.Getenv("GAME_FB_EXPOSE_CALLBACKS") == "true" {
		fbCallback := r.Group("/fb/callback")
		{
			fbCallback.POST("/health", fb_api.CallbackHealth)
			fbCallback.POST("/balance", fb_api.CallbackBalance)
			fbCallback.POST("/order_pay", fb_api.CallbackOrderPay)
			fbCallback.POST("/check_order_pay", fb_api.CallbackCheckOrderPay)
			fbCallback.POST("/sync_transaction", fb_api.CallbackSyncTransaction)
			fbCallback.POST("/sync_orders", fb_api.CallbackSyncOrders)
			fbCallback.POST("/sync_cashout", fb_api.CallbackSyncCashout)
		}
	}

	if os.Getenv("GAME_SABA_EXPOSE_CALLBACKS") == "true" {
		sabaCallback := r.Group("/saba")
		{
			sabaCallback.POST("/getbalance", saba_api.CallbackGetBalance)
			sabaCallback.POST("/placebet", saba_api.CallbackPlaceBet)
			sabaCallback.POST("/confirmbet", saba_api.CallbackConfirmBet)
			sabaCallback.POST("/cancelbet", saba_api.CallbackCancelBet)
			sabaCallback.POST("/settle", saba_api.CallbackSettle)
			sabaCallback.POST("/unsettle", saba_api.CallbackUnsettle)
			sabaCallback.POST("/resettle", saba_api.CallbackResettle)
			sabaCallback.POST("/placebetparlay", saba_api.CallbackPlaceBetParlay)
			sabaCallback.POST("/confirmbetparlay", saba_api.CallbackConfirmBetParlay)
		}
	}

	if os.Getenv("GAME_DC_EXPOSE_CALLBACKS") == "true" {
		dcCallback := r.Group("/dcs")
		{
			dcCallback.POST("/login", dc_api.CallbackLogin)
			dcCallback.POST("/getBalance", dc_api.CallbackLogin) // getBalance response is the same as login
			dcCallback.POST("/wager", dc_api.CallbackWager)
			dcCallback.POST("/cancelWager", dc_api.CallbackCancelWager)
			dcCallback.POST("/appendWager", dc_api.CallbackAppendWager)
			dcCallback.POST("/endWager", dc_api.CallbackEndWager)
			dcCallback.POST("/freeSpinResult", dc_api.CallbackFreeSpinResult)
			dcCallback.POST("/promoPayout", dc_api.CallbackPromoPayout)
		}
	}

	if os.Getenv("GAME_IMSB_EXPOSE_CALLBACKS") == "true" {
		imsbCallback := r.Group("/imsb")
		{
			imsbCallback.GET("/ValidateToken", imsb_api.ValidateToken)
			imsbCallback.GET("/GetBalance", imsb_api.GetBalance)
			imsbCallback.GET("/GetApproval", imsb_api.GetBalance) // same as GetBalance
			imsbCallback.POST("/DeductBalance", imsb_api.DeductBalance)
			imsbCallback.POST("/UpdateBalance", imsb_api.UpdateBalance)
		}
	}

	if os.Getenv("GAME_DOLLAR_JACKPOT_EXPOSE_CALLBACKS") == "true" {
		djCallback := r.Group("/dollar_jackpot")
		{
			djCallback.POST("/settle_order", dollar_jackpot_api.SettleOrder)
		}
	}

	if os.Getenv("FINPAY_CALLBACK_ENABLED") == "true" {
		fpCallback := r.Group("/callback/finpay")
		fpCallback.POST("/payment-order", middleware.RequestLogger("Finpay callback"), api_finpay.FinpayPaymentCallback)
		fpCallback.POST("/transfer-order", middleware.RequestLogger("Finpay callback"), api_finpay.FinpayTransferCallback)
	}

	internal := r.Group("/internal")
	{
		internal.Use(middleware.BrandAgent())
		internal.POST("/top-up-order/close", middleware.RequestLogger("internal"), internal_api.FinpayBackdoor)
		internal.POST("/withdraw-order/reject", middleware.RequestLogger("internal"), internal_api.RejectWithdrawal)
		internal.POST("/withdraw-order/approve", middleware.RequestLogger("internal"), internal_api.ApproveWithdrawal)
		internal.POST("/withdraw-order/insert", middleware.RequestLogger("internal"), internal_api.CustomOrder)
	}

	// middlewares order can't be changed
	r.Use(middleware.Cors())
	r.GET("/ts", api.Ts)
	captcha := r.Group("/captcha")
	{
		captcha.POST("/get", api.CaptchaGet)
		captcha.POST("/check", api.CaptchaCheck)
	}

	r.Use(middleware.EncryptPayload())
	r.Use(middleware.CheckSignature())
	r.Use(middleware.Ip())
	r.Use(middleware.BrandAgent())
	r.Use(middleware.Timezone())
	r.Use(middleware.Location())
	r.Use(middleware.Locale())

	r.GET("/ping", api.Ping)

	v1 := r.Group("/v1")
	{
		// no cache routes
		v1.POST("/sms_otp", api.SmsOtp)
		v1.POST("/email_otp", api.EmailOtp)
		v1.POST("/login_otp", api.UserLoginOtp)
		v1.POST("/login_password", api.UserLoginPassword)
		v1.POST("/password", middleware.CheckAuth(), api.UserSetPassword)
		v1.GET("/otp-check", api.VerifyOtp)

		v1.GET("/config", middleware.Cache(10*time.Minute), api.Config)
		v1.GET("/app_update", middleware.Cache(1*time.Minute), api.AppUpdate)
		v1.GET("/announcements", middleware.CheckAuth(), middleware.CacheForGuest(1*time.Minute), api.Announcements)
		v1.GET("/categories", middleware.Cache(1*time.Minute), api.CategoryList)
		v1.GET("/vendors", middleware.Cache(1*time.Minute), api.VendorList)
		v1.GET("/streams", middleware.Cache(1*time.Minute), api.StreamList)
		v1.GET("/streamer", middleware.Cache(1*time.Minute), api.Streamer)
		v1.GET("/topup-methods", middleware.CheckAuth(), api.TopupMethodList)
		v1.GET("/withdraw-methods", middleware.CheckAuth(), api.WithdrawMethodList)
		v1.GET("/avatars", middleware.Cache(1*time.Minute), api.AvatarList)
		v1.POST("/share", api.ShareCreate)
		v1.GET("/share", api.ShareGet)
		v1.GET("/games", middleware.Cache(1*time.Minute), api.GameList)
		v1.GET("/room_chat/history", api.RoomChatHistory)

		v1.GET("/promotion/list", middleware.CheckAuth(), middleware.Cache(5*time.Minute), promotion_api.GetCoverList)
		v1.GET("/promotion/details", middleware.CheckAuth(), middleware.CacheForGuest(5*time.Minute), promotion_api.GetDetail)

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
			dj.GET("/", dollar_jackpot_api.DollarJackpotGet) // can't have cache due to "total" value; cache at the query
			dj.GET("/winners", dollar_jackpot_api.DollarJackpotWinners)
		}

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
				user.POST("/nickname", api.NicknameUpdate)
				user.POST("/profile_pic", api.ProfilePicUpload)
				user.GET("/notifications", api.UserNotificationList)
				user.PUT("/notification/mark_read", api.UserNotificationMarkRead)
				user.GET("/counters", api.UserCounters)
				user.PUT("/fcm_token", api.FcmTokenUpdate)

				user.GET("/following_ids", api.UserFollowingIdList)
				user.POST("/follow", api.UserFollowingAdd)
				user.DELETE("/follow", api.UserFollowingRemove)

				user.GET("/favourites", api.UserFavouriteList)
				user.POST("/favourite", api.UserFavouriteAdd)
				user.DELETE("/favourite", api.UserFavouriteRemove)

				user.PUT("/secondary-password", api.UserSetSecondaryPassword)

				user.GET("/orders", api.OrderList)
				user.GET("/recent_games", api.UserRecentGameList)

				user.POST("/withdraw-accounts", api.WthdrawAccountsAdd)
				user.DELETE("/withdraw-accounts", api.WthdrawAccountsRemove)
				user.GET("/withdraw-accounts", api.WthdrawAccountsList)

				user.GET("/otp-check", api.VerifyOtp)

				user.POST("/transfer_to", api.TransferTo)
				user.POST("/transfer_from", api.TransferFrom)
				user.POST("/transfer_back", api.TransferBack)

				user.POST("/feedback", api.FeedbackAdd)

				taya := user.Group("/taya")
				{
					taya.GET("/token", taya_api.GetToken)
				}

				fb := user.Group("/fb")
				{
					fb.GET("/token", fb_api.GetToken)
				}

				dc := user.Group("/dc")
				{
					dc.GET("/get_url", dc_api.GetUrl)
				}

				imsb := user.Group("/imsb")
				{
					imsb.GET("/token", imsb_api.GetToken)
					imsb.POST("/apply_voucher", imsb_api.ApplyVoucher)
				}

				dj := user.Group("/dollar_jackpot")
				{
					dj.POST("/place_order", dollar_jackpot_api.PlaceOrder)
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
					promotion.GET("/list", middleware.Cache(5*time.Minute), promotion_api.GetCoverList)
					promotion.GET("/details", middleware.RequestLogger("get promotion details"), promotion_api.GetDetail)
					promotion.POST("/claim", middleware.RequestLogger("promotion claim"), promotion_api.PromotionClaim)
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
			}
		}
		v1.GET("/user/heartbeat", middleware.AuthRequired(false, false), api.Heartbeat)
	}

	return r
}
