package dc

import (
	"web-api/model"
	"web-api/service/common"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/dc/callback"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
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
func (c *FreeSpinResult) GetBetAmountOnly() int64 {
	return 0
}
func (c *FreeSpinResult) GetExternalUserId() string {
	return c.Request.BrandUid
}

func (c *FreeSpinResult) GetBetAmount() (amount int64, exists bool) {
	return
}

func FreeSpinResultCallback(c *gin.Context, req callback.FreeSpinResultRequest) (res callback.BaseResponse, err error) {
	go common.LogGameCallbackRequest("free_spin_result", req)

	cl := util.DCFactory.NewClient()
	err = cl.VerifySign(req)
	if err != nil {
		res = SignErrorResponse()
		return
	}
	
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
