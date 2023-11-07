package api

import (
	"github.com/gin-gonic/gin"
	"web-api/service"
)

func Streamer(c *gin.Context) {
	var service service.StreamerService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.Get(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
