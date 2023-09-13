package server

import (
	"os"
	"time"
	"web-api/api"
	fb_api "web-api/api/fb"
	saba_api "web-api/api/saba"
	"web-api/middleware"

	"github.com/gin-gonic/gin"
)

// NewRouter 路由配置
func NewRouter() *gin.Engine {
	r := gin.New()
	r.GET("/ts", api.Ts)

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
		}
	}

	// 中间件, 顺序不能改
	r.Use(middleware.Cors())
	r.Use(middleware.CheckSignature())
	r.Use(middleware.Ip())
	r.Use(middleware.BrandAgent())
	r.Use(middleware.Timezone())
	r.Use(middleware.Location())
	r.Use(middleware.Locale())
	r.Use(middleware.EncryptPayload())
	r.Use(middleware.CorrelationID())
	r.Use(middleware.ErrorLogStatus())

	r.GET("/ping", api.Ping)

	v1 := r.Group("/v1")
	{
		// no cache routes
		v1.POST("/sms_otp", api.SmsOtp)
		v1.POST("/email_otp", api.EmailOtp)
		v1.POST("/login_otp", api.UserLoginOtp)
		v1.POST("/login_password", api.UserLoginPassword)
		v1.POST("/password", middleware.CheckAuth(), api.UserSetPassword)

		v1.GET("/config", middleware.Cache(10*time.Minute), api.Config)
		v1.GET("/categories", middleware.Cache(1*time.Minute), api.CategoryList)
		v1.GET("/streams", middleware.Cache(1*time.Minute), api.StreamList)
		v1.GET("/streamer", middleware.Cache(1*time.Minute), api.Streamer)
		v1.GET("/cash-methods", api.CashMethodList)


		saba := v1.Group("/saba")
		{
			saba.GET("/get_url", middleware.CheckAuth(), saba_api.GetUrl)
		}

		auth := v1.Group("")
		auth.Use(middleware.AuthRequired(true))
		{
			user := auth.Group("/user")
			{
				user.GET("/me", api.Me)
				user.DELETE("/me", api.UserDelete)
				user.DELETE("/logout", api.UserLogout)
				user.POST("/finish_setup", api.UserFinishSetup)
				user.GET("/check_username", api.UserCheckUsername)
				user.GET("/check_password", api.UserCheckPassword)

				fb := user.Group("/fb")
				{
					fb.GET("/token", fb_api.GetToken)
				}

				kyc := user.Group("/kyc")
				{
					kyc.GET("", api.GetKyc)
					kyc.POST("", api.SubmitKyc)
				}
			}

		}
		v1.GET("/user/heartbeat", middleware.AuthRequired(false), api.Heartbeat)
	}

	return r
}
