package server

import (
	"web-api/api"
	internal_api "web-api/api/internal_api"
	"web-api/middleware"

	"github.com/gin-gonic/gin"
)

func InitInternalApi(r *gin.Engine) {
	internal := r.Group("/internal")
	{
		internal.Use(middleware.BrandAgent())
		internal.POST("/top-up-order/close", middleware.RequestLogger("internal"), internal_api.FinpayBackdoor)
		internal.POST("/withdraw-order/reject", middleware.RequestLogger("internal"), internal_api.RejectWithdrawal)
		internal.POST("/withdraw-order/approve", middleware.RequestLogger("internal"), internal_api.ApproveWithdrawal)
		internal.POST("/withdraw-order/insert", middleware.RequestLogger("internal"), internal_api.CustomOrder)
		internal.POST("/withdraw-order/manual-approve", middleware.RequestLogger("internal"), internal_api.ManualCloseCashOut)
		internal.PUT("/recall", middleware.RequestLogger("internal"), api.InternalRecallFund)
		internal.POST("/promotions/batch-claim", middleware.RequestLogger("internal"), internal_api.InternalPromotion)
		internal.POST("/promotions/custom", middleware.RequestLogger("internal"), internal_api.InternalPromotionRequest)
		internal.POST("/notification-push", middleware.RequestLogger("internal"), internal_api.Notification)
		internal.POST("/notification-push-all", middleware.RequestLogger("internal"), internal_api.NotificationAll)
	}
}
