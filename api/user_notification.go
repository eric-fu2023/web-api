package api

import (
	"web-api/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func UserNotificationList(c *gin.Context) {
	var service service.UserNotificationListService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.List(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserNotificationMarkRead(c *gin.Context) {
	var service service.UserNotificationMarkReadService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res, _ := service.MarkRead(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserNotificationListV2(c *gin.Context) {
	var service service.UserNotificationListService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.List(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}