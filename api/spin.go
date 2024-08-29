package api

import (
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func SpinHistory(c *gin.Context) {
	var service service.SpinService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.GetHistory(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}