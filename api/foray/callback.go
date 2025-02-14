package api_foray

import (
	cashin_foray "web-api/service/cashin/foray"
	cashout_foray "web-api/service/cashout/foray"
	"web-api/util"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func ForayPaymentCallback(c *gin.Context) {
	var service cashin_foray.ForayPaymentCallback
	if err := c.ShouldBindWith(&service, binding.JSON); err == nil {
		if err := service.Handle(c); err == nil {
			c.String(200, "ok")
		} else {
			util.GetLoggerEntry(c).Error(err)
			c.String(500, "failed")
		}
	} else {
		c.String(400, "param error: "+err.Error())
	}
}

func ForayTransferCallback(c *gin.Context) {
	var service cashout_foray.ForayTransferCallback
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
