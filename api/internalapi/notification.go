package internal_api

import (
	"web-api/serializer"
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func Notification(c *gin.Context) {
	var service service.InternalNotificationPushRequest
	if err := c.ShouldBind(&service); err == nil {
		res := service.Handle(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, serializer.ParamErr(c, service, "", err))
	}
}

func NotificationAll(c *gin.Context) {
	var service service.InternalNotificationPushAllRequest
	if err := c.ShouldBind(&service); err == nil {
		res := service.HandleAll(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, serializer.ParamErr(c, service, "", err))
	}
}