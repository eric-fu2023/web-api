package api_finpay

import (
	cashin_finpay "web-api/service/cashin/finpay"
	cashout_finpay "web-api/service/cashout/finpay"
	"web-api/util"

	"github.com/gin-gonic/gin"
)

func FinpayPaymentCallback(c *gin.Context) {
	var service cashin_finpay.FinpayPaymentCallback
	if err := c.ShouldBind(&service); err == nil {
		if err := service.Handle(c); err == nil {
			c.String(200, "success")
		} else {
			util.GetLoggerEntry(c).Error(err)
			c.String(500, "failed")
		}
	} else {
		c.String(400, "param error")
	}
}

func FinpayTransferCallback(c *gin.Context) {
	var service cashout_finpay.FinpayTransferCallback
	if err := c.ShouldBind(&service); err == nil {
		if err := service.Handle(c); err == nil {
			c.String(200, "success")
		} else {
			util.GetLoggerEntry(c).Error(err)
			c.String(500, "failed")
		}
	} else {
		c.String(400, "param error")
	}
}
