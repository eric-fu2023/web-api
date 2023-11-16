package saba

import (
	"blgit.rfdev.tech/taya/game-service/saba/callback"
	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/service/common"
)

type ConfirmBet struct {
	Callback
	Request       callback.ConfirmBetRequest
	ChangedAmount int64
}

func (c *ConfirmBet) NewCallback(userId int64) {
	for _, txn := range c.Request.Message.Txns {
		var existingTx models.SabaTransaction
		rows := model.DB.Where(`ref_id`, txn.RefId).First(&existingTx).RowsAffected
		if rows == 0 {
			continue
		}
		c.Transaction = existingTx
		c.Transaction.CfmOperationId = c.Request.Message.OperationId
		if v, e := time.Parse(time.RFC3339, c.Request.Message.UpdateTime); e == nil {
			t := v.UTC()
			c.Transaction.CfmUpdateTime = &t
		}
		if v, e := time.Parse(time.RFC3339, c.Request.Message.TransactionTime); e == nil {
			t := v.UTC()
			c.Transaction.CfmTransactionTime = &t
		}
		c.Transaction.CfmTxId = txn.TxId
		c.Transaction.CfmIsOddsChanged = txn.IsOddsChanged
		if v, e := time.Parse(time.RFC3339, txn.WinlostDate); e == nil {
			t := v.UTC()
			c.Transaction.CfmWinlostDate = &t
		}
		c.Transaction.ActualAmount = int64(txn.ActualAmount * 100)
		c.ChangedAmount = int64(txn.CreditAmount * 100)
		c.Transaction.DebitAmount = c.Transaction.DebitAmount - c.ChangedAmount
	}
}

func (c *ConfirmBet) GetExternalUserId() string {
	return c.Request.Message.UserId
}

func (c *ConfirmBet) ShouldProceed() bool {
	if c.Transaction.RefId == "" { // if confirmbet refId doesn't match with any of the existing refId
		return false
	}
	return true
}

func (c *ConfirmBet) GetAmount() int64 {
	return c.ChangedAmount
}

func (c *ConfirmBet) GetBetAmount() (amount int64, exists bool) {
	return
}

func (c *ConfirmBet) IsAdjustment() bool {
	return true
}

func ConfirmBetCallback(c *gin.Context, req callback.ConfirmBetRequest) (res any, err error) {
	go common.LogGameCallbackRequest("confirmbet", req)
	clb := ConfirmBet{Request: req}
	err = common.ProcessTransaction(&clb)
	if err != nil {
		return
	}
	_, balance, _, _, err := common.GetUserAndSum(model.DB, consts.GameVendor["saba"], req.Message.UserId)
	if err != nil {
		return
	}
	res = callback.ConfirmCancelBetResponse{
		BaseResponse: callback.BaseResponse{
			Status: "0",
		},
		Balance: float64(balance) / 100,
	}
	return
}
