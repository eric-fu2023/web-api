package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"web-api/service"
)

func GetKyc(c *gin.Context) {
	var service service.GetKycService
	if err := c.ShouldBind(&service); err == nil {
		res := service.GetKyc(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func SubmitKyc(c *gin.Context) {
	var service service.SubmitKycService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res := service.SubmitKyc(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
