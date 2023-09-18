package api

import (
	"github.com/gin-gonic/gin"
	"web-api/service"
)

func Config(c *gin.Context) {
	var service service.AppConfigService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func Announcements(c *gin.Context) {
	var service service.AnnouncementsService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
