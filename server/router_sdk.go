package server

import (
	"os"

	"web-api/api"
	dc_api "web-api/api/dc"
	dollar_jackpot_api "web-api/api/dollar_jackpot"
	fb_api "web-api/api/fb"
	api_finpay "web-api/api/finpay"
	api_foray "web-api/api/foray"
	imsb_api "web-api/api/imsb"
	saba_api "web-api/api/saba"
	stream_game_api "web-api/api/stream_game"
	taya_api "web-api/api/taya"
	"web-api/middleware"

	"github.com/gin-gonic/gin"
)

func InitSdkApi(r *gin.Engine) {
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

	if os.Getenv("GAME_STREAM_GAME_EXPOSE_CALLBACKS") == "true" {
		djCallback := r.Group("/stream_game")
		{
			djCallback.POST("/settle_order", stream_game_api.SettleOrder)
		}
	}

	/* payment channel callback */
	if os.Getenv("FINPAY_CALLBACK_ENABLED") == "true" {
		fpCallback := r.Group("/callback/finpay")
		fpCallback.POST("/payment-order", middleware.RequestLogger("Finpay callback"), api_finpay.FinpayPaymentCallback)
		fpCallback.POST("/transfer-order", middleware.RequestLogger("Finpay callback"), api_finpay.FinpayTransferCallback)
	}
	if os.Getenv("FORAY_CALLBACK_ENABLED") == "true" {
		fpCallback := r.Group("/callback/foray")
		fpCallback.POST("/payment-order", middleware.RequestLogger("Foray callback"), api_foray.ForayPaymentCallback)
		fpCallback.POST("/transfer-order", middleware.RequestLogger("Foray callback"), api_foray.ForayTransferCallback)
	}

	if os.Getenv("BACKEND_APIS_ON") == "true" {
		backend := r.Group("/backend")
		{
			backend.POST("/token", api.BackendGetToken)
			backend.POST("/pin", api.BackendSetPin)
		}
	}
}
