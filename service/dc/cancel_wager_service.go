package dc

import (
	"blgit.rfdev.tech/taya/game-service/dc/callback"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"web-api/model"
	"web-api/service/common"
)

type CancelWager struct {
	Callback
	Request     callback.CancelWagerRequest
	WagerExists bool
}

func (c *CancelWager) NewCallback(userId int64) {
	copier.Copy(&c.Transaction, &c.Request)
	c.Transaction.UserId = userId
	c.Transaction.ExternalUserId = c.Request.BrandUid
	if amount, exists := c.Callback.GetBetAmount(); exists {
		c.WagerExists = exists
		c.Transaction.Amount = amount
	}
}

func (c *CancelWager) GetAmount() int64 {
	var multiplier int64 = 1
	if c.Transaction.WagerType == 2 { // 1: cancelWager (should credit amount to user); 2: cancelEndWager (should debit amount from user)
		multiplier = -1
	}
	return multiplier * c.Transaction.Amount
}

func (c *CancelWager) GetExternalUserId() string {
	return c.Request.BrandUid
}

func (c *CancelWager) ShouldProceed() bool {
	return c.WagerExists
}

func CancelWagerCallback(c *gin.Context, req callback.CancelWagerRequest) (res callback.BaseResponse, err error) {
	j, _ := json.Marshal(req)
	fmt.Println(`cancel_wager: `, string(j))
	res, err = CheckDuplicate(c, model.ByDcRoundWagerAndWagerType(req.RoundId, req.WagerId), req.BrandUid)
	if res.Code != 0 || err != nil {
		return
	}

	a := CancelWager{Request: req}
	err = common.ProcessTransaction(&a)
	if err != nil {
		return
	}
	res, err = SuccessResponse(c, req.BrandUid)
	return
}
