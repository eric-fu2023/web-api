package saba_api

import (
	"blgit.rfdev.tech/taya/game-service/saba/callback"
	"github.com/gin-gonic/gin"
	"os"
	"web-api/api"
	"web-api/service/saba"
	"web-api/util"
)

func CallbackGetBalance(c *gin.Context) {
	compressRequest(c)
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
	compressRequest(c)
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
	compressRequest(c)
	var req callback.ConfirmBetRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := saba.ConfirmBetCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackCancelBet(c *gin.Context) {
	compressRequest(c)
	var req callback.CancelBetRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := saba.CancelBetCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackSettle(c *gin.Context) {
	compressRequest(c)
	var req callback.SettleRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := saba.SettleCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackUnsettle(c *gin.Context) {
	compressRequest(c)
	var req callback.SettleRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := saba.UnsettleCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackResettle(c *gin.Context) {
	compressRequest(c)
	var req callback.SettleRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := saba.ResettleCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackPlaceBetParlay(c *gin.Context) {
	compressRequest(c)
	var req callback.PlaceBetParlayRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := saba.PlaceBetParlayCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackConfirmBetParlay(c *gin.Context) {
	compressRequest(c)
	var req callback.ConfirmBetParlayRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := saba.ConfirmBetParlayCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func compressRequest(c *gin.Context) {
	if os.Getenv("ENV") == "local" {
		return
	}
	decompressedBody, e := callback.DecompressRequest(c.Request.Body)
	if e != nil {
		c.JSON(200, ErrorResponse(c, nil, e))
		return
	}
	c.Request.Body = decompressedBody
}

func ErrorResponse(c *gin.Context, req any, err error) (res callback.BaseResponse) {
	res = callback.BaseResponse{
		Status: "203",
		Msg:    err.Error(),
	}
	util.Log().Error(res.Msg, c.Request.URL, c.Request.Header, util.MarshalService(req))
	return
}
