package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"web-api/service"
)

func CsHistory(c *gin.Context) {
	var service service.CsHistoryService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func CsSend(c *gin.Context) {
	var service service.CsSendService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res, _ := service.Send(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
