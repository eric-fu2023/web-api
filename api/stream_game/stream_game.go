package stream_game_api

import (
	"github.com/gin-gonic/gin"
	"web-api/api"
	"web-api/service/stream_game"
)

func StreamGame(c *gin.Context) {
	var service stream_game.StreamGameService
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

func StreamGameList(c *gin.Context) {
	var service stream_game.StreamGameServiceList
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
