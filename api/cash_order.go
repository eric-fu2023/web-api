package api

import (
	"web-api/serializer"
	"web-api/service"
	"web-api/service/cashin"
	"web-api/service/cashout"

	"github.com/gin-gonic/gin"
)

func TopUpOrder(c *gin.Context) {
	var service cashin.TopUpOrderService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.CreateOrder(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(200, serializer.EnsureErr(c, err, res))
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func WithdrawOrder(c *gin.Context) {
	var service cashout.WithdrawOrderService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Do(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(200, serializer.EnsureErr(c, err, res))
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func ListCashOrder(c *gin.Context) {
	var service service.CashOrderService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.List(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
