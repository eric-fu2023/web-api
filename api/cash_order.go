package api

import (
	"web-api/service"
	"web-api/service/cashin"
	"web-api/service/cashout"

	"github.com/gin-gonic/gin"
)

func TopUpOrder(c *gin.Context) {
	var service cashin.TopUpOrderService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.CreateOrder(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func WithdrawOrder(c *gin.Context) {
	var service cashout.WithdrawOrderService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Do(c)
		c.JSON(200, res)
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