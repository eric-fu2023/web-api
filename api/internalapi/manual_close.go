package internal_api

import (
	cashin_finpay "web-api/service/cashin/finpay"

	"github.com/gin-gonic/gin"
)

func FinpayBackdoor(c *gin.Context) {
	var service cashin_finpay.ManualCloseService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Do(c); err == nil {
			c.String(200, "success")
		} else {
			c.JSON(500, res)
		}
	} else {
		c.String(400, "param error")
	}
}
