package fb_api

import (
	"blgit.rfdev.tech/taya/game-service/fb/callback"
	"github.com/gin-gonic/gin"
	"web-api/api"
	"web-api/serializer"
	"web-api/service/fb"
)

func CallbackHealth(c *gin.Context) {
	c.JSON(200, serializer.Response{
		Code: 0,
	})
}

func CallbackBalance(c *gin.Context) {
	var req callback.BalanceRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := fb.BalanceCallback(c, req); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackOrderPay(c *gin.Context) {
	var req callback.OrderPayRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := fb.OrderPayCallback(c, req); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}