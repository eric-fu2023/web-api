package api

import (
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func WinLose(c *gin.Context) {
	var service service.CsHistoryService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func Vip(c *gin.Context) {
	var service service.CsHistoryService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func SpinItems(c *gin.Context) {
	var service service.SpinService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}