package internal_api

import (
	"web-api/api"
	cashin_finpay "web-api/service/cashin/finpay"

	"github.com/gin-gonic/gin"
)

func FinpayBackdoor(c *gin.Context) {
	var service cashin_finpay.ManualCloseService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Do(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, api.ErrorResponse(c, service, err))
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
