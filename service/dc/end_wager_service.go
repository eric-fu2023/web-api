package dc

import (
	"web-api/model"
	"web-api/service/common"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/dc/callback"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
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
	go common.LogGameCallbackRequest("end_wager", req)

	cl := util.DCFactory.NewClient()
	err = cl.VerifySign(req)
	if err != nil {
		res = SignErrorResponse()
		return
	}

	res, err = CheckRound(c, req.RoundId, req.BrandUid)
	if res.Code != 0 || err != nil {
		return
	}

	res, err = CheckDuplicate(c, model.ByDcRoundAndWager(req.RoundId, req.WagerId), req.BrandUid)
	if res.Code != 0 || err != nil {
		return
	}

	a := EndWager{Request: req}
	err = common.ProcessTransaction(&a)
	if err != nil {
		return
	}
	res, err = SuccessResponse(c, req.BrandUid)
	return
}
