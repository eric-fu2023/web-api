package fb_api

import (
	"blgit.rfdev.tech/taya/game-service/fb/callback"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"web-api/api"
	"web-api/service/fb"
)

func CallbackHealth(c *gin.Context) {
	c.JSON(200, callback.BaseResponse{
		Code: 0,
	})
}

func CallbackBalance(c *gin.Context) {
	var req callback.BalanceRequest
	if err := c.ShouldBind(&req); err == nil {
		res, _ := fb.BalanceCallback(c, req)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackOrderPay(c *gin.Context) {
	var req callback.OrderPayRequest
	if err := c.ShouldBind(&req); err == nil {
		res, _ := fb.OrderPayCallback(c, req)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackCheckOrderPay(c *gin.Context) {
	var req callback.OrderPayRequest
	if err := c.ShouldBind(&req); err == nil {
		res, _ := fb.CheckOrderPayCallback(c, req)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackSyncTransaction(c *gin.Context) {
	var req []callback.OrderPayRequest
	if err := c.ShouldBind(&req); err == nil {
		res, _ := fb.SyncTransactionCallback(c, req)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackSyncOrders(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	fmt.Println(string(body))
	c.JSON(200, callback.BaseResponse{
		Code: 0,
	})
}