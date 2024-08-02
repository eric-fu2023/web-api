package taya_api

import (
	"blgit.rfdev.tech/taya/game-service/fb/callback"
	"github.com/gin-gonic/gin"
	"log"
	"web-api/api"
	"web-api/service/taya"
	"web-api/util"
)

func CallbackHealth(c *gin.Context) {
	c.JSON(200, callback.BaseResponse{
		Code: 0,
	})
}

func CallbackBalance(c *gin.Context) {
	log.Printf("taya/fb CallbackBalance...")
	var req callback.BalanceRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := taya.BalanceCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackOrderPay(c *gin.Context) {
	var req callback.OrderPayRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := taya.OrderPayCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackCheckOrderPay(c *gin.Context) {
	var req callback.OrderPayRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := taya.CheckOrderPayCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackSyncTransaction(c *gin.Context) {
	var req []callback.OrderPayRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := taya.SyncTransactionCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackSyncOrders(c *gin.Context) {
	var req callback.SyncOrdersRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := taya.SyncOrdersCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackSyncCashout(c *gin.Context) {
	var req callback.SyncCashoutRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := taya.SyncCashoutCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func ErrorResponse(c *gin.Context, req any, err error) (res callback.BaseResponse) {
	res = callback.BaseResponse{
		Code:    1,
		Message: err.Error(),
	}
	util.Log().Error(res.Message, c.Request.URL, c.Request.Header, util.MarshalService(req))
	return
}
