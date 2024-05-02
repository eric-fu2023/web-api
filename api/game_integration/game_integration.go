package game_integration_api

import (
	"web-api/api"
	"web-api/service/game_integration"

	"github.com/gin-gonic/gin"
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

func GameCategoryList(c *gin.Context) {
	var service game_integration.GameCategoryListService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.List(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func SubGames(c *gin.Context) {
	var service game_integration.SubGameService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.List(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
