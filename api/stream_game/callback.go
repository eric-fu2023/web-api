package stream_game_api

import (
	"github.com/gin-gonic/gin"
	"web-api/api"
	"web-api/service/stream_game"
	"web-api/util"
)

func PlaceOrder(c *gin.Context) {
	var req stream_game.PlaceOrder
	if err := c.ShouldBind(&req); err == nil {
		res, e := stream_game.Place(c, req)
		c.JSON(200, res)
		if e != nil {
			util.Log().Error("stream game place order: ", e)
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func SettleOrder(c *gin.Context) {
	var req stream_game.SettleOrder
	if err := c.ShouldBind(&req); err == nil {
		res, e := stream_game.Settle(c, req)
		c.JSON(200, res)
		if e != nil {
			util.Log().Error("stream game settle order: ", e)
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}
