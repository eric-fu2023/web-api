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

type FreeSpinResult struct {
	Callback
	Request callback.FreeSpinResultRequest
}

func (c *FreeSpinResult) NewCallback(userId int64) {
	copier.Copy(&c.Transaction, &c.Request)
	c.Transaction.UserId = userId
	c.Transaction.ExternalUserId = c.Request.BrandUid
	c.Transaction.Amount = int64(c.Request.Amount * 100)
}

func (c *FreeSpinResult) GetAmount() int64 {
	return c.Transaction.Amount
}

func (c *FreeSpinResult) GetExternalUserId() string {
	return c.Request.BrandUid
}

func (c *FreeSpinResult) GetBetAmount() (amount int64, exists bool) {
	return
}

func FreeSpinResultCallback(c *gin.Context, req callback.FreeSpinResultRequest) (res callback.BaseResponse, err error) {
	j, _ := json.Marshal(req)
	fmt.Println(`free_spin_result: `, string(j))
	res, err = CheckDuplicate(c, model.ByDcRoundAndWager(req.RoundId, req.WagerId), req.BrandUid)
	if res.Code != 0 || err != nil {
		return
	}

	a := FreeSpinResult{Request: req}
	err = common.ProcessTransaction(&a)
	if err != nil {
		return
	}
	res, err = SuccessResponse(c, req.BrandUid)
	return
}
