package api

import (
	"github.com/gin-gonic/gin"
	"web-api/service/backend"
)

func BackendGetToken(c *gin.Context) {
	var service backend.GetTokenService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func BackendSetPin(c *gin.Context) {
	var service backend.PinService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Set(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
