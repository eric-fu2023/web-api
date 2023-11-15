package dc_api

import (
	"blgit.rfdev.tech/taya/game-service/dc/callback"
	"github.com/gin-gonic/gin"
	"web-api/api"
	"web-api/service/dc"
	"web-api/util"
)

func CallbackLogin(c *gin.Context) {
	var req callback.LoginRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := dc.SuccessResponseWithTokenCheck(c, req.BrandUid, req.Token); err != nil {
			c.JSON(200, ErrorResponse(c, req, res.Code, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackWager(c *gin.Context) {
	var req callback.WagerRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := dc.WagerCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, res.Code, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackCancelWager(c *gin.Context) {
	var req callback.CancelWagerRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := dc.CancelWagerCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, res.Code, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackAppendWager(c *gin.Context) {
	var req callback.AppendWagerRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := dc.AppendWagerCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, res.Code, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackEndWager(c *gin.Context) {
	var req callback.EndWagerRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := dc.EndWagerCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, res.Code, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackFreeSpinResult(c *gin.Context) {
	var req callback.FreeSpinResultRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := dc.FreeSpinResultCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, res.Code, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func CallbackPromoPayout(c *gin.Context) {
	var req callback.PromoPayoutRequest
	if err := c.ShouldBind(&req); err == nil {
		if res, err := dc.PromoPayoutCallback(c, req); err != nil {
			c.JSON(200, ErrorResponse(c, req, res.Code, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func ErrorResponse(c *gin.Context, req any, code int64, err error) (res callback.BaseResponse) {
	var codeNumber int64 = 1001
	if code > 1000 {
		codeNumber = code
	}
	res = callback.BaseResponse{
		Code: codeNumber,
		Msg:  err.Error(),
	}
	util.Log().Error(res.Msg, c.Request.URL, c.Request.Header, util.MarshalService(req))
	return
}
