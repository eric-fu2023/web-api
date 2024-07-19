package api

import (
	"web-api/serializer"
	"web-api/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func GiftSend(c *gin.Context) {
	var service service.GiftSendRequestService
	if err := c.ShouldBindWith(&service, binding.Form); err == nil {
		res, err := service.Handle(c)
		if err != nil {
			c.JSON(400, serializer.ParamErr(c, service, err.Error(), err))
			return
		}
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func GiftRecordList(c *gin.Context) {
	var service service.GiftRecordListService
	if err := c.ShouldBindWith(&service, binding.Form); err == nil {
		res, err := service.List(c)
		if err != nil {
			c.JSON(500, serializer.Err(c, "", 50000, err.Error(), err))
			return
		}
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
