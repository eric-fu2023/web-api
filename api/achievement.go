package api

import (
	"github.com/gin-gonic/gin"
	"web-api/service"
)

func AchievementComplete(c *gin.Context) {
	var service service.AchievementCompleteService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.Complete(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
