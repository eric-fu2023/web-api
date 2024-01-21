package imsb

import (
	"blgit.rfdev.tech/taya/game-service/imsb/callback"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/service/common"
)

func GetBalanceCallback(c *gin.Context, req callback.GetBalanceRequest, dateReceived string) (res callback.CommonWalletBaseResponse, err error) {
	_, balance, _, _, err := common.GetUserAndSum(model.DB, consts.GameVendor["imsb"], req.MemberCode)
	if err != nil {
		return
	}
	res = callback.CommonWalletBaseResponse{
		PackageId:     0,
		Balance:       float64(balance) / 100,
		DateReceived:  strings.Replace(dateReceived, "%20", "T", 1),
		DateSent:      time.Now().Format("2006-01-02T15:04:05"),
		StatusCode:    100,
		StatusMessage: "Acknowledge",
	}
	return
}
