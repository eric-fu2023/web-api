package saba

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/service/common"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/saba/callback"
	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

type Settle struct {
	Callback
	OperationId string
	Request     callback.SettleTxns
}

func (c *Settle) NewCallback(userId int64) {
	var existingTx models.SabaTransaction
	rows := model.DB.Where(`ref_id`, c.Request.RefId).Where(`status != ''`).First(&existingTx).RowsAffected
	if rows != 0 {
		c.Transaction = existingTx
	}
	copier.Copy(&c.Transaction, &c.Request)
	if v, e := time.Parse(time.RFC3339, c.Request.UpdateTime); e == nil {
		c.Transaction.UpdateTime = v.UTC()
	}
	if v, e := time.Parse(time.RFC3339, c.Request.WinlostDate); e == nil {
		t := v.UTC()
		c.Transaction.WinlostDate = &t
	}
	c.Transaction.OperationId = c.OperationId
	c.Transaction.Payout = int64(c.Request.Payout * 100)
	c.Transaction.ActualAmount = int64(c.Request.Payout * 100)
	if c.Request.DebitAmount != 0 {
		c.Transaction.DebitAmount = c.Transaction.CreditAmount - int64(c.Request.DebitAmount*100)
	}
	c.Transaction.CreditAmount = int64(c.Request.CreditAmount * 100)
}

func (c *Settle) GetExternalUserId() string {
	return c.Request.UserId
}

func (c *Settle) ShouldProceed() bool {
	return true
}

func (c *Settle) GetBetAmountOnly() int64 {
	return 0
}
type SettleRedis struct {
	OpId string
	Txn  callback.SettleTxns
}

func SettleCallback(c *gin.Context, req callback.SettleRequest) (res any, err error) {
	go common.LogGameCallbackRequest("settle", req)
	go func(c *gin.Context, req callback.SettleRequest) {
		for _, txn := range req.Message.Txns {
			sr := SettleRedis{OpId: req.Message.OperationId, Txn: txn}
			jj, _ := json.Marshal(sr)
			_, e := cache.RedisSyncTransactionClient.Set(context.TODO(), fmt.Sprintf(`saba:%s:%d`, txn.UserId, txn.TxId), jj, 0).Result()
			if e != nil {
				util.Log().Error("redis error", e)
			}
		}
	}(c, req)
	res = callback.BaseResponse{
		Status: "0",
	}

	return
}

func UnsettleCallback(c *gin.Context, req callback.SettleRequest) (res any, err error) {
	go common.LogGameCallbackRequest("unsettle", req)
	go func(c *gin.Context, req callback.SettleRequest) {
		for _, txn := range req.Message.Txns {
			txn.Status = "unsettle"
			txn.Payout = 0
			sr := SettleRedis{OpId: req.Message.OperationId, Txn: txn}
			jj, _ := json.Marshal(sr)
			_, e := cache.RedisSyncTransactionClient.Set(context.TODO(), fmt.Sprintf(`saba:%s:%d`, txn.UserId, txn.TxId), jj, 0).Result()
			if e != nil {
				util.Log().Error("redis error", e)
			}
		}
	}(c, req)
	res = callback.BaseResponse{
		Status: "0",
	}

	return
}

func ResettleCallback(c *gin.Context, req callback.SettleRequest) (res any, err error) {
	go common.LogGameCallbackRequest("resettle", req)
	go func(c *gin.Context, req callback.SettleRequest) {
		for _, txn := range req.Message.Txns {
			sr := SettleRedis{OpId: req.Message.OperationId, Txn: txn}
			jj, _ := json.Marshal(sr)
			_, e := cache.RedisSyncTransactionClient.Set(context.TODO(), fmt.Sprintf(`saba:%s:%d`, txn.UserId, txn.TxId), jj, 0).Result()
			if e != nil {
				util.Log().Error("redis error", e)
			}
		}
	}(c, req)
	res = callback.BaseResponse{
		Status: "0",
	}

	return
}
