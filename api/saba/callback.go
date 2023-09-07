package saba_api

import (
	"blgit.rfdev.tech/taya/game-service/saba/callback"
	"github.com/gin-gonic/gin"
	"web-api/api"
	"web-api/service/saba"
	"web-api/util"
)

func CallbackGetBalance(c *gin.Context) {
	decompressedBody, e := callback.DecompressRequest(c.Request.Body)
	if e != nil {
		c.JSON(200, ErrorResponse(c, nil, e))
		return
	}
	c.Request.Body = decompressedBody
	var req callback.GetBalanceRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := saba.GetBalanceCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackPlaceBet(c *gin.Context) {
	decompressedBody, e := callback.DecompressRequest(c.Request.Body)
	if e != nil {
		c.JSON(200, ErrorResponse(c, nil, e))
		return
	}
	c.Request.Body = decompressedBody
	var req callback.PlaceBetRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := saba.PlaceBetCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackConfirmBet(c *gin.Context) {
	c.JSON(200, map[string]interface{}{
		"status":  "0",
		"balance": "1001.20",
	})
}

func ErrorResponse(c *gin.Context, req any, err error) (res callback.BaseResponse) {
	res = callback.BaseResponse{
		Status: "203",
		Msg:    err.Error(),
	}
	util.Log().Error(res.Msg, c.Request.URL, c.Request.Header, util.MarshalService(req))
	return
}
