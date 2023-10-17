package api

import (
	"github.com/gin-gonic/gin"
	"web-api/service"
)

func ShareCreate(c *gin.Context) {
	var service service.CreateShareService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Create(); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, res)
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func ShareGet(c *gin.Context) {
	var service service.GetShareService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Get(); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, res)
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
