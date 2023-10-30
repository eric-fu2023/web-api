package dc

import (
	"blgit.rfdev.tech/taya/game-service/dc/callback"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"web-api/model"
	"web-api/service/common"
)

type Wager struct {
	Callback
	Request callback.WagerRequest
}

func (c *Wager) NewCallback(userId int64) {
	copier.Copy(&c.Transaction, &c.Request)
	c.Transaction.UserId = userId
	c.Transaction.ExternalUserId = c.Request.BrandUid
	c.Transaction.Amount = int64(c.Request.Amount * 100)
}

func (c *Wager) GetAmount() int64 {
	return -1 * c.Transaction.Amount
}

func (c *Wager) GetExternalUserId() string {
	return c.Request.BrandUid
}

func WagerCallback(c *gin.Context, req callback.WagerRequest) (res callback.BaseResponse, err error) {
	go common.LogGameCallbackRequest("wager", req)
	res, err = CheckDuplicate(c, model.ByDcRoundAndWager(req.RoundId, req.WagerId), req.BrandUid)
	if res.Code != 0 || err != nil {
		return
	}

	a := Wager{Request: req}
	err = common.ProcessTransaction(&a)
	if err != nil {
		return
	}
	res, err = SuccessResponse(c, req.BrandUid)
	return
}
