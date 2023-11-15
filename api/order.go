package api

import (
	"github.com/gin-gonic/gin"
	"web-api/service"
)

func OrderList(c *gin.Context) {
	var service service.OrderListService
	if err := c.ShouldBind(&service); err == nil {
		res := service.List(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
