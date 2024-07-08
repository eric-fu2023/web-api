package internal_api

import (
	"web-api/serializer"
	cashout_finpay "web-api/service/cashout/finpay"

	"github.com/gin-gonic/gin"
)

func RejectWithdrawal(c *gin.Context) {
	var service cashout_finpay.CashOutOrderService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Reject(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, serializer.EnsureErr(c, err, res))
		}
	} else {
		c.JSON(400, serializer.ParamErr(c, service, "", err))
	}
}

func ApproveWithdrawal(c *gin.Context) {
	var service cashout_finpay.CashOutOrderService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Approve(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, serializer.Err(c, "", 50000, err.Error(), err))
		}
	} else {
		c.JSON(400, serializer.ParamErr(c, service, "", err))
	}
}
