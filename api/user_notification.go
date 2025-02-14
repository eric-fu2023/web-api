package api

import (
	"web-api/service"
	notificationservice "web-api/service/notification"

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
	var service service.UserNotificationListServiceV2
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.List(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserNotificationCategoryV2(c *gin.Context) {
	var service service.UserNotificationListCategoryServiceV2
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.ListCategory(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func GeneralNotificationV2(c *gin.Context) {
	var req service.GetGeneralNotificationRequestV2
	if err := c.ShouldBind(&req); err == nil {
		res, _ := service.GetGeneralNotificationV2(c, req)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, req, err))
	}
}

func AddReadNotificationsV2(c *gin.Context) {
	var req notificationservice.UserNotificationMarkReadRequestV2
	if err := c.ShouldBind(&req); err == nil {
		res, _ := service.AddReadNotificationsV2(c, req)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, req, err))
	}
}
