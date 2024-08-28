package api

import (
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func CheckOrder(c *gin.Context) {
	var service service.CheckOrderService
	if err := c.ShouldBind(&service); err == nil {
		res := service.CheckOrder(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func OrderList(c *gin.Context) {
	var service service.OrderListService
	if err := c.ShouldBind(&service); err == nil {
		res := service.List(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
