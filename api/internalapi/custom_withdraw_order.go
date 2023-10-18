package internal_api

import (
	"web-api/api"
	"web-api/service/cashout"

	"github.com/gin-gonic/gin"
)

func CustomOrder(c *gin.Context) {
	var service cashout.CustomOrderService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Handle(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, api.ErrorResponse(c, service, err))
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
