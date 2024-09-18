package internal_api

import (
	"web-api/serializer"
	"web-api/service/cashout"

	"github.com/gin-gonic/gin"
)

func CustomOrder(c *gin.Context) {
	var service cashout.CustomOrderService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Handle(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, serializer.EnsureErr(c, err, res))
		}
	} else {
		c.JSON(400, serializer.ParamErr(c, service, "", err))
	}
}

func ManualCloseCashOut(c *gin.Context) {
	var service cashout.ManualCloseOrderService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Do(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, serializer.EnsureErr(c, err, res))
		}
	} else {
		c.JSON(400, serializer.ParamErr(c, service, "", err))
	}
}
