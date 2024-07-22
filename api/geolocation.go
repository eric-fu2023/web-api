package api

import (
	"web-api/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func GeolocationCreate(c *gin.Context) {
	var service service.CreateGeolocationService
	if err := c.ShouldBindWith(&service, binding.JSON); err == nil {
		res := service.Create(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func GeolocationGet(c *gin.Context) {
	var service service.GetGeolocationService
	if err := c.ShouldBind(&service); err == nil {
		res := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
