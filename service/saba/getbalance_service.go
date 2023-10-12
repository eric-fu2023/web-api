package saba

import (
	"blgit.rfdev.tech/taya/game-service/saba/callback"
	"github.com/gin-gonic/gin"
	"time"
	"web-api/conf/consts"
	"web-api/service/common"
)

func GetBalanceCallback(c *gin.Context, req callback.GetBalanceRequest) (res any, err error) {
	gpu, balance, _, _, err := common.GetUserAndSum(consts.GameVendor["saba"], req.Message.UserId)
	if err != nil {
		return
	}
	now := time.Now().In(time.FixedZone("GMT-4", -4*60*60))
	res = callback.GetBalanceResponse{
		BaseResponse: callback.BaseResponse{
			Status: "0",
		},
		UserId:    gpu.ExternalUserId,
		Balance:   float64(balance) / 100,
		BalanceTs: now.Format(time.RFC3339),
	}
	return
}
