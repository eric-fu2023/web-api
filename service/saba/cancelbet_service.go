package saba

import (
	"blgit.rfdev.tech/taya/game-service/saba/callback"
	models "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/service"
)

type CancelBet struct {
	Callback
	Request       callback.CancelBetRequest
	ChangedAmount int64
}

func (c *CancelBet) NewCallback(userId int64) {
	for _, txn := range c.Request.Message.Txns {
		var existingTx models.SabaTransactionC
		rows := model.DB.Where(`ref_id`, txn.RefId).First(&existingTx).RowsAffected
		if rows == 0 {
			continue
		}
		c.Transaction = existingTx
		c.Transaction.CancOperationId = c.Request.Message.OperationId
		if v, e := time.Parse(time.RFC3339, c.Request.Message.UpdateTime); e == nil {
			t := v.UTC()
			c.Transaction.CancUpdateTime = &t
		}
		ChangedAmount := int64(txn.CreditAmount * 100)
		c.Transaction.ActualAmount = c.Transaction.ActualAmount - ChangedAmount
		c.Transaction.DebitAmount = c.Transaction.DebitAmount - ChangedAmount
	}
}

func (c *CancelBet) GetExternalUserId() string {
	return c.Request.Message.UserId
}

func (c *CancelBet) ShouldProceed() bool {
	if c.Transaction.RefId == "" {
		return false
	}
	return true
}

func (c *CancelBet) GetAmount() int64 {
	return c.ChangedAmount
}

func (c *CancelBet) IsAdjustment() bool {
	return true
}

func CancelBetCallback(c *gin.Context, req callback.CancelBetRequest) (res any, err error) {
	j, _ := json.Marshal(req)
	fmt.Println("cancelbet: ", string(j))
	for i, _ := range req.Message.Txns {
		r := req
		r.Message.Txns = r.Message.Txns[i : i+1]
		clb := CancelBet{Request: r}
		err = service.ProcessTransaction(&clb)
		if err != nil {
			return
		}
	}
	_, balance, _, _, err := service.GetUserAndSum(consts.GameProvider["saba"], req.Message.UserId)
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
