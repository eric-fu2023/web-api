package saba

import (
	"blgit.rfdev.tech/taya/game-service/saba/callback"
	models "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/service/fb"
)

func GetBalanceCallback(c *gin.Context, req callback.GetBalanceRequest) (res any, err error) {
	gpu, err := fb.GetGameProviderUser(consts.GameProvider["saba"], req.Message.UserId)
	if err != nil {
		return
	}

	balance, _, _, err := fb.GetSums(gpu)
	if err != nil {
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

func PlaceBetCallback(c *gin.Context, req callback.PlaceBetRequest) (res any, err error) {
	j, _ := json.Marshal(req)
	fmt.Println("placebet: ", string(j))
	gpu, err := fb.GetGameProviderUser(consts.GameProvider["saba"], req.Message.UserId)
	if err != nil {
		return
	}
	var tx models.SabaTransactionC
	copier.Copy(&tx, &req.Message)
	if v, e := time.Parse(time.RFC3339, req.Message.KickOffTime); e == nil {
		tx.KickOffTime = v.UTC()
	}
	if v, e := time.Parse(time.RFC3339, req.Message.BetTime); e == nil {
		tx.BetTime = v.UTC()
	}
	if v, e := time.Parse(time.RFC3339, req.Message.UpdateTime); e == nil {
		tx.UpdateTime = v.UTC()
	}
	if v, e := time.Parse(time.RFC3339, req.Message.MatchDatetime); e == nil {
		tx.MatchDatetime = v.UTC()
	}
	tx.UserId = gpu.UserId
	tx.ExternalUserId = req.Message.UserId
	tx.BetAmount = req.Message.BetAmount * 100
	tx.ActualAmount = req.Message.ActualAmount * 100
	tx.CreditAmount = req.Message.CreditAmount * 100
	tx.DebitAmount = req.Message.DebitAmount * 100
	err = model.DB.Save(&tx).Error
	if err != nil {
		return
	}

	res = callback.PlaceBetResponse{
		BaseResponse: callback.BaseResponse{
			Status: "0",
		},
		RefId:        req.Message.RefId,
		LicenseeTxId: tx.ID,
	}
	return
}
