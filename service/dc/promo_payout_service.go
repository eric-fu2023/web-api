package dc

import (
	"web-api/service/common"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/dc/callback"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

type PromoPayout struct {
	Callback
	Request callback.PromoPayoutRequest
}

func (c *PromoPayout) NewCallback(userId int64) {
	copier.Copy(&c.Transaction, &c.Request)
	c.Transaction.UserId = userId
	c.Transaction.ExternalUserId = c.Request.BrandUid
	c.Transaction.Amount = int64(c.Request.Amount * 100)
}

func (c *PromoPayout) GetAmount() int64 {
	return c.Transaction.Amount
}

func (c *PromoPayout) GetExternalUserId() string {
	return c.Request.BrandUid
}

func (c *PromoPayout) GetBetAmount() (amount int64, exists bool) {
	return
}

func PromoPayoutCallback(c *gin.Context, req callback.PromoPayoutRequest) (res callback.BaseResponse, err error) {
	go common.LogGameCallbackRequest("promo_payout", req)

	cl := util.DCFactory.NewClient()
	err = cl.VerifySign(req)
	if err != nil {
		res = SignErrorResponse()
		return
	}
	
	a := PromoPayout{Request: req}
	err = common.ProcessTransaction(&a)
	if err != nil {
		return
	}
	res, err = SuccessResponse(c, req.BrandUid)
	return
}
