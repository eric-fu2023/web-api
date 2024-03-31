package game_integration_api

import (
	"github.com/gin-gonic/gin"
	"web-api/api"
	"web-api/service/game_integration"
)

func GetUrl(c *gin.Context) {
	var service game_integration.GetUrlService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.Get(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
