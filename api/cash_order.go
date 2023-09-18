package api

import (
	"web-api/service/cashin"

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