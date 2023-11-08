package api

import (
	"github.com/gin-gonic/gin"
	"web-api/service"
)

func CaptchaGet(c *gin.Context) {
	var service service.CaptchaGetService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func CaptchaCheck(c *gin.Context) {
	var service service.CaptchaCheckService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Check(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
