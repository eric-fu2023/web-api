package dc

import (
	"blgit.rfdev.tech/taya/game-service/dc/callback"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/service/common"
)

type Callback struct {
	Transaction ploutos.DcTransactionC
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

func (c *Callback) GetAmount() int64 {
	return c.Transaction.Amount
}

func (c *Callback) GetWagerMultiplier() (int64, bool) {
	return -1, true
}

func (c *Callback) GetBetAmount() (amount int64, exists bool) {
	e := model.DB.Model(ploutos.DcTransactionC{}).Select(`amount`).
		Where(`wager_id`, c.Transaction.WagerId).Order(`id`).First(&amount).Error
	if e == nil {
		exists = true
	}
	return
}

func (c *Callback) IsAdjustment() bool {
	return false
}

func SuccessResponse(c *gin.Context, brandUid string) (res callback.BaseResponse, err error) {
	gpu, balance, _, _, err := common.GetUserAndSum(consts.GameVendor["dc"], brandUid)
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

func CheckDuplicate(c *gin.Context, scope func(*gorm.DB) *gorm.DB, brandUid string) (res callback.BaseResponse, err error) {
	var dcTx ploutos.DcTransactionC
	rows := model.DB.Model(ploutos.DcTransactionC{}).Scopes(scope).First(&dcTx).RowsAffected
	if rows > 0 {
		res, err = DuplicatedTxResponse(c, brandUid)
	}
	return
}

func DuplicatedTxResponse(c *gin.Context, brandUid string) (res callback.BaseResponse, err error) {
	gpu, balance, _, _, err := common.GetUserAndSum(consts.GameVendor["dc"], brandUid)
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
