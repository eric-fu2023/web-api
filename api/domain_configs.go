package api

import (
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func DomainInitApp(c *gin.Context) {
	var service service.DomainConfigService
	if err := c.ShouldBind(&service); err == nil {
		code, res, err := service.InitApp(c)
		if err == nil {
			c.JSON(code, res)
		} else {
			c.JSON(400, ErrorResponse(c, service, err))
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
