package fb_api

import (
	"blgit.rfdev.tech/taya/game-service/fb/callback"
	"github.com/gin-gonic/gin"
	"log"
	"web-api/api"
	"web-api/service/fb"
	"web-api/util"
)

func CallbackHealth(c *gin.Context) {
	c.JSON(200, callback.BaseResponse{
		Code: 0,
	})
}

func CallbackBalance(c *gin.Context) {
	var req callback.BalanceRequest

	log.Printf("CallbackBalance... \n")
	if err := c.ShouldBind(&req); err == nil {
		if res, err := fb.BalanceCallback(c, req); err != nil {
			log.Printf("CallbackBalance fb.BalanceCallback() err %v \n", err)
			c.JSON(200, ErrorResponse(c, req, err))
		} else {
			log.Printf("CallbackBalance fb.BalanceCallback() ok %v \n", err)
			c.JSON(200, res)
		}
	} else {
		log.Printf("CallbackBalance... err %v \n", err)
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackOrderPay(c *gin.Context) {
	var req callback.OrderPayRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := fb.OrderPayCallback(c, req); err != nil {
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
		if res, err := fb.CheckOrderPayCallback(c, req); err != nil {
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
		if res, err := fb.SyncTransactionCallback(c, req); err != nil {
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
		if res, err := fb.SyncOrdersCallback(c, req); err != nil {
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
		if res, err := fb.SyncCashoutCallback(c, req); err != nil {
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
