package dc

import (
	"context"
	"errors"
	"fmt"
	"time"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/service/common"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/dc/callback"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Callback struct {
	Transaction ploutos.DcTransaction
}

func (c *Callback) GetGameVendorId() int64 {
	return consts.GameVendor["dc"]
}

func (c *Callback) GetGameTransactionId() int64 {
	return c.Transaction.ID
}

func (c *Callback) SaveGameTransaction(tx *gorm.DB) error {
	return tx.Save(&c.Transaction).Error
}

func (c *Callback) ShouldProceed() bool {
	return true // dc doesn't have wager that shouldn't proceed
}

func (c *Callback) GetWagerMultiplier() (int64, bool) {
	return -1, true
}

func (c *Callback) GetBetAmount() (amount int64, exists bool) {
	e := model.DB.Model(ploutos.DcTransaction{}).Select(`amount`).
		Where(`round_id`, c.Transaction.RoundId).Where(`bet_type`, 1).Order(`id`).First(&amount).Error
	if e == nil {
		exists = true
	}
	return
}

func (c *Callback) IsAdjustment() bool {
	return false
}

func SuccessResponse(c *gin.Context, brandUid string) (res callback.BaseResponse, err error) {
	gpu, balance, _, _, err := common.GetUserAndSum(model.DB, consts.GameVendor["dc"], brandUid)
	if err != nil {
		return
	}
	res = callback.BaseResponse{
		Code: 1000,
		Data: callback.CommonResponse{
			BrandUid: gpu.ExternalUserId,
			Currency: gpu.ExternalCurrency,
			Balance:  float64(balance) / 100,
		},
	}
	return
}

func SuccessResponseWithTokenCheck(c *gin.Context, req callback.LoginRequest) (res callback.BaseResponse, err error) {
	cl := util.DCFactory.NewClient()
	err = cl.VerifySign(req)
	if err != nil {
		res = SignErrorResponse()
		return
	}
	res, err = CheckToken(req.BrandUid, req.Token)
	if res.Code != 0 || err != nil {
		return
	}
	res, err = SuccessResponse(c, req.BrandUid)
	return
}

func CheckDuplicate(c *gin.Context, scope func(*gorm.DB) *gorm.DB, brandUid string) (res callback.BaseResponse, err error) {
	var dcTx ploutos.DcTransaction
	rows := model.DB.Model(ploutos.DcTransaction{}).Scopes(scope).First(&dcTx).RowsAffected
	if rows > 0 {
		res, err = DuplicatedTxResponse(c, brandUid)
	}
	return
}

func CheckRound(c *gin.Context, roundId string, brandUid string) (res callback.BaseResponse, err error) {
	var dcTx ploutos.DcTransaction
	rows := model.DB.Model(ploutos.DcTransaction{}).Where("round_id", roundId).First(&dcTx).RowsAffected
	if rows == 0 {
		res, err = MissingRoundResponse(c, brandUid)
	}
	return
}

func DuplicatedTxResponse(c *gin.Context, brandUid string) (res callback.BaseResponse, err error) {
	gpu, balance, _, _, err := common.GetUserAndSum(model.DB, consts.GameVendor["dc"], brandUid)
	if err != nil {
		return
	}
	res = callback.BaseResponse{
		Code: 5043,
		Data: callback.CommonResponse{
			BrandUid: gpu.ExternalUserId,
			Currency: gpu.ExternalCurrency,
			Balance:  float64(balance) / 100,
		},
	}
	return
}

func MissingRoundResponse(c *gin.Context, brandUid string) (res callback.BaseResponse, err error) {
	gpu, balance, _, _, err := common.GetUserAndSum(model.DB, consts.GameVendor["dc"], brandUid)
	if err != nil {
		return
	}
	res = callback.BaseResponse{
		Code: 5042,
		Data: callback.CommonResponse{
			BrandUid: gpu.ExternalUserId,
			Currency: gpu.ExternalCurrency,
			Balance:  float64(balance) / 100,
		},
	}
	return
}

func InsufficientBalanceResponse(c *gin.Context, brandUid string) (res callback.BaseResponse, err error) {
	gpu, balance, _, _, err := common.GetUserAndSum(model.DB, consts.GameVendor["dc"], brandUid)
	if err != nil {
		return
	}
	res = callback.BaseResponse{
		Code: 5003,
		Data: callback.CommonResponse{
			BrandUid: gpu.ExternalUserId,
			Currency: gpu.ExternalCurrency,
			Balance:  float64(balance) / 100,
		},
	}
	return
}

func CheckToken(brandUid string, token string) (res callback.BaseResponse, err error) {
	redisKey := fmt.Sprintf(`hacksaw_token:%s:%s`, brandUid, token)
	r := cache.RedisSessionClient.Get(context.TODO(), redisKey)
	if r.Err() == redis.Nil {
		res = TokenErrorResponse()
		err = errors.New("token invalid")
		return
	}
	go func() {
		cache.RedisSessionClient.Expire(context.TODO(), redisKey, 2*time.Hour)
	}()
	return
}

func TokenErrorResponse() (res callback.BaseResponse) {
	res = callback.BaseResponse{
		Code: 5009,
	}
	return
}

func SignErrorResponse() (res callback.BaseResponse) {
	res = callback.BaseResponse{
		Code: 5000,
	}
	return
}
