package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"web-api/service"
)

func TransferTo(c *gin.Context) {
	var service service.TransferToService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res, e := service.Do(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func TransferFrom(c *gin.Context) {
	var service service.TransferFromService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res, e := service.Do(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func TransferBack(c *gin.Context) {
	var service service.TransferBackService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res, e := service.Do(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
