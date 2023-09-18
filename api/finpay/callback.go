package finpay

import (
	cashin_finpay "web-api/service/cashin/finpay"
	"web-api/util"

	"github.com/gin-gonic/gin"
)

func FinpayCallback(c *gin.Context) {
	var service cashin_finpay.FinpayCallback
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
