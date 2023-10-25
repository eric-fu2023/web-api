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

type EndWager struct {
	Callback
	Request callback.EndWagerRequest
}

func (c *EndWager) NewCallback(userId int64) {
	copier.Copy(&c.Transaction, &c.Request)
	c.Transaction.UserId = userId
	c.Transaction.ExternalUserId = c.Request.BrandUid
	c.Transaction.Amount = int64(c.Request.Amount * 100)
}

func (c *EndWager) GetAmount() int64 {
	return c.Transaction.Amount
}

func (c *EndWager) GetExternalUserId() string {
	return c.Request.BrandUid
}

func EndWagerCallback(c *gin.Context, req callback.EndWagerRequest) (res callback.BaseResponse, err error) {
	j, _ := json.Marshal(req)
	fmt.Println(`end_wager: `, string(j))
	res, err = CheckDuplicate(c, model.ByDcRoundAndWager(req.RoundId, req.WagerId), req.BrandUid)
	if res.Code != 0 || err != nil {
		return
	}

	a := EndWager{Request: req}
	err = common.ProcessTransaction(&a)
	if err != nil {
		return
	}
	res, err = SuccessResponse2(c, req.BrandUid)
	return
}
