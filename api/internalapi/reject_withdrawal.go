package internal_api

import (
	"web-api/api"
	cashout_finpay "web-api/service/cashout/finpay"

	"github.com/gin-gonic/gin"
)

func RejectWithdrawal(c *gin.Context) {
	var service cashout_finpay.CancelCashOutOrderService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Reject(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func ApproveWithdrawal(c *gin.Context) {
	var service cashout_finpay.ApproveCashOutOrderService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Approve(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
