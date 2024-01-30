package dc

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/plugin/dbresolver"
	"strings"
	"time"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/service/common"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/dc/callback"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
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
	e := model.DB.Clauses(dbresolver.Use("txConn")).Model(ploutos.DcTransaction{}).Select(`amount`).
		Where(`round_id`, c.Transaction.RoundId).Where(`bet_type`, 1).Order(`id`).First(&amount).Error
	if e == nil {
		exists = true
	}
	return
}

func (c *Callback) IsAdjustment() bool {
	return false
}

func (c *Callback) ApplyInsuranceVoucher(userId int64, betAmount int64, betExists bool) (err error) {
	// Voucher application not done
	return
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

func CheckRound(c *gin.Context, roundId string, wagerId string, brandUid string) (res callback.BaseResponse, err error) {
	var dcTx ploutos.DcTransaction
	q := model.DB.Model(ploutos.DcTransaction{}).Where("round_id", roundId)
	if wagerId != "" {
		q = q.Where(`wager_id`, wagerId)
	}
	rows := q.First(&dcTx).RowsAffected
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
	redisKey := fmt.Sprintf(`hacksaw_token:*:%s`, token)
	ctx := context.TODO()
	iter := cache.RedisSessionClient.Scan(ctx, 0, redisKey, 0).Iterator()
	var n int64
	var notOwnToken bool
	for iter.Next(ctx) {
		n++
		redisKey = iter.Val()
		str := strings.Split(iter.Val(), ":")
		if str[1] != brandUid {
			notOwnToken = true
			break
		}
	}
	if n == 0 {
		res = NonExistenceTokenErrorResponse()
		err = errors.New("token invalid")
		return
	} else if notOwnToken {
		res = NotOwnTokenErrorResponse()
		err = errors.New("token invalid")
		return
	}
	go func() {
		cache.RedisSessionClient.Expire(context.TODO(), redisKey, 2*time.Hour)
	}()
	return
}

func NotOwnTokenErrorResponse() (res callback.BaseResponse) {
	res = callback.BaseResponse{
		Code: 5013,
	}
	return
}

func NonExistenceTokenErrorResponse() (res callback.BaseResponse) {
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
