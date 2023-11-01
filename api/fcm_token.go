package api

import (
	"github.com/gin-gonic/gin"
	"web-api/service"
)

func FcmTokenUpdate(c *gin.Context) {
	var service service.FcmTokenUpdateService
	if err := c.ShouldBind(&service); err == nil {
		res := service.Update(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
