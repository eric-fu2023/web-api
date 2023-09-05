package saba

import (
	"blgit.rfdev.tech/taya/game-service/saba/callback"
	"github.com/gin-gonic/gin"
	"time"
	"web-api/conf/consts"
	"web-api/service/fb"
	"web-api/util"
)

func GetBalanceCallback(c *gin.Context, req callback.GetBalanceRequest) (res any, err error) {
	gpu, err := fb.GetGameProviderUser(consts.GameProvider["saba"], req.Message.UserId)
	if err != nil {
		res = callbackErrorResponse(c, req, err)
		return
	}

	balance, _, _, err := fb.GetSums(gpu)
	if err != nil {
		res = callbackErrorResponse(c, req, err)
		return
	}

	now := time.Now().In(time.FixedZone("GMT-4", -4*60*60))
	res = callback.GetBalanceResponse{
		BaseResponse: callback.BaseResponse{
			Status: "0",
		},
		UserId:    req.Message.UserId,
		Balance:   float64(balance) / 100,
		BalanceTs: now.Format(time.RFC3339),
	}
	return
}

func callbackErrorResponse(c *gin.Context, req any, err error) (res callback.BaseResponse) {
	res = callback.BaseResponse{
		Status: "203",
		Msg:    err.Error(),
	}
	util.Log().Error(res.Msg, c.Request.URL, c.Request.Header, util.MarshalService(req))
	return
}
