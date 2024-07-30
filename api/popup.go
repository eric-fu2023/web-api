package api

import (
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func WinLose(c *gin.Context) {
	var service service.WinLoseService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
func WinLoseShown(c *gin.Context) {
	var service service.WinLoseService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Shown(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func Vip(c *gin.Context) {
	var service service.VipService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func VipShown(c *gin.Context) {
	var service service.VipService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Shown(c)
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

func SpinResult(c *gin.Context) {
	var service service.SpinService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Result(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}