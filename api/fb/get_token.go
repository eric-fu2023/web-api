package fb_api

import (
	"github.com/gin-gonic/gin"
	"web-api/api"
	"web-api/service/fb"
)

func GetToken(c *gin.Context) {
	var service fb.TokenService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Get(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
