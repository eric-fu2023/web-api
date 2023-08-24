package server

import (
	"github.com/gin-gonic/gin"
	"os"
	"web-api/api"
	fb_api "web-api/api/fb"
	"web-api/middleware"
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

	r.GET("/ping", api.Ping)

	v1 := r.Group("/v1")
	{
		// no cache routes
		v1.POST("/sms_otp", api.SmsOtp)
		v1.POST("/email_otp", api.EmailOtp)
		v1.POST("/login_otp", api.UserLoginOtp)
		v1.POST("/login_password", api.UserLoginPassword)

		auth := v1.Group("")
		auth.Use(middleware.AuthRequired())
		{
			user := auth.Group("/user")
			{
				user.GET("/me", api.Me)
				user.DELETE("/me", api.UserDelete)
				user.DELETE("/logout", api.UserLogout)
				user.POST("/password", api.UserSetPassword)

				fb := user.Group("/fb")
				{
					fb.GET("/token", fb_api.GetToken)
				}
			}
		}
	}

	return r
}
