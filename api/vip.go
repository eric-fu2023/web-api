package api

import (
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func VipGet(c *gin.Context) {
	var service service.VipQuery
	if err := c.ShouldBind(&service); err == nil {
		res := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func VipLoad(c *gin.Context) {
	var service service.VipLoad
	if err := c.ShouldBind(&service); err == nil {
		res := service.Load(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
