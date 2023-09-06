package saba_api

import (
	"blgit.rfdev.tech/taya/game-service/saba/callback"
	"github.com/gin-gonic/gin"
	"web-api/api"
	"web-api/service/saba"
)

func CallbackGetBalance(c *gin.Context) {
	decompressedBody, _ := callback.DecompressRequest(c.Request.Body)
	c.Request.Body = decompressedBody
	var req callback.GetBalanceRequest
	if err := c.ShouldBind(&req); err == nil {
		res, _ := saba.GetBalanceCallback(c, req)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}
