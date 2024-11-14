package api

import (
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func Pixel(c *gin.Context) {
	var service service.PixelInstall
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.PixelInstallEvent(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

