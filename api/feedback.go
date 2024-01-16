package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"web-api/service"
)

func FeedbackAdd(c *gin.Context) {
	var service service.FeedbackAddService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res, _ := service.Add(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
