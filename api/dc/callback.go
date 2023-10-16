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
		if res, err := dc.SuccessResponse(c, req.BrandUid); err != nil {
			c.JSON(200, ErrorResponse(c, req, err))
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
		Code: 1001,
		Msg:  err.Error(),
	}
	util.Log().Error(res.Msg, c.Request.URL, c.Request.Header, util.MarshalService(req))
	return
}
