package imsb

import (
	"blgit.rfdev.tech/taya/game-service/imsb/callback"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"strings"
	"time"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/service/common"
	"web-api/util"
)

type Callback struct {
	Request     callback.WagerDetail
	Transaction ploutos.ImTransaction
}

func (c *Callback) NewCallback(userId int64) {
	c.Transaction.UserId = userId
	c.Transaction.ActionId = c.Request.ActionId
	c.Transaction.ExternalUserId = c.Request.MemberCode
	c.Transaction.WagerNo = c.Request.WagerNo
	c.Transaction.TransactionAmount = int64(c.Request.TransactionAmount * 100)
}

func (c *Callback) GetGameVendorId() int64 {
	return consts.GameVendor["imsb"]
}

func (c *Callback) GetGameTransactionId() int64 {
	return c.Transaction.ID
}

func (c *Callback) GetExternalUserId() string {
	return c.Request.MemberCode
}

func (c *Callback) SaveGameTransaction(tx *gorm.DB) error {
	return tx.Save(&c.Transaction).Error
}

func (c *Callback) ShouldProceed() bool {
	return true // imsb tx should always proceed
}

func (c *Callback) GetAmount() int64 {
	return c.Transaction.TransactionAmount
}

func (c *Callback) GetWagerMultiplier() (value int64, exists bool) {
	return -1, true
}

func (c *Callback) GetBetAmount() (amount int64, exists bool) {
	e := model.DB.Clauses(dbresolver.Use("txConn")).Model(ploutos.ImTransaction{}).Select(`transaction_amount`).
		Where(`wager_no`, c.Transaction.WagerNo).Where(`action_id`, 0).Order(`id`).First(&amount).Error
	if e == nil {
		exists = true
	}
	return
}

func (c *Callback) IsAdjustment() bool {
	return false
}

func (c *Callback) ApplyInsuranceVoucher(userId int64, betAmount int64, betExists bool) (err error) {
	// Voucher application not done
	return
}

func GetBalanceCallback(c *gin.Context, req callback.GetBalanceRequest, enc callback.EncryptedRequest) (res callback.CommonWalletBaseResponse, err error) {
	_, balance, _, _, err := common.GetUserAndSum(model.DB, consts.GameVendor["imsb"], req.MemberCode)
	if err != nil {
		return
	}
	res = callback.CommonWalletBaseResponse{
		PackageId:     enc.PackageId,
		Balance:       float64(balance) / 100,
		DateReceived:  strings.Replace(enc.DateSent, "%20", "T", 1),
		DateSent:      time.Now().Format("2006-01-02T15:04:05"),
		StatusCode:    100,
		StatusMessage: "Acknowledge",
	}
	return
}

func DeductBalanceCallback(c *gin.Context, req callback.WagerDetail, enc callback.EncryptedRequest) (res callback.CommonWalletBaseResponse, err error) {
	go common.LogGameCallbackRequest("DeductBalance", req)
	_, balance, _, _, err := common.GetUserAndSum(model.DB, consts.GameVendor["imsb"], req.MemberCode)
	if err != nil {
		return
	}
	duplicate := CheckDuplicate(req.WagerNo)
	if !duplicate {
		err = common.ProcessTransaction(&Callback{Request: req})
		if err != nil {
			return
		}
	}
	res = callback.CommonWalletBaseResponse{
		PackageId:     enc.PackageId,
		Balance:       float64(balance) / 100,
		DateReceived:  strings.Replace(enc.DateSent, "%20", "T", 1),
		DateSent:      time.Now().Format("2006-01-02T15:04:05"),
		StatusCode:    100,
		StatusMessage: "Acknowledge",
	}
	return
}

func UpdateBalanceCallback(c *gin.Context, req callback.UpdateBalanceRequest, enc callback.EncryptedRequest) (res callback.CommonWalletBaseResponse, err error) {
	go common.LogGameCallbackRequest("UpdateBalance", req)
	go func(c *gin.Context, req callback.UpdateBalanceRequest) {
		for _, r := range req.WagerDetails {
			r.ActionId = req.ActionId
			j, _ := json.Marshal(r)
			_, e := cache.RedisSyncTransactionClient.Set(context.TODO(), fmt.Sprintf(`im:%s:%s`, r.MemberCode, r.WagerNo), j, 0).Result()
			if e != nil {
				util.Log().Error("imsb UpdateBalance insert into redis error: ", e, r)
			}
		}
	}(c, req)
	res = callback.CommonWalletBaseResponse{
		PackageId:     enc.PackageId,
		DateReceived:  strings.Replace(enc.DateSent, "%20", "T", 1),
		DateSent:      time.Now().Format("2006-01-02T15:04:05"),
		StatusCode:    100,
		StatusMessage: "Acknowledge",
	}
	return
}

func CheckDuplicate(wagerNo string) bool {
	var imTx ploutos.ImTransaction
	rows := model.DB.Model(ploutos.ImTransaction{}).Where(`wager_no`, wagerNo).First(&imTx).RowsAffected
	if rows > 0 {
		return true
	}
	return false
}
