package api

import (
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func GeolocationGet(c *gin.Context) {
	var service service.GetGeolocationService
	if err := c.ShouldBind(&service); err == nil {
		res := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
