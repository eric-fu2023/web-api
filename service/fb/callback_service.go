package fb

import (
	"blgit.rfdev.tech/taya/game-service/fb/callback"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"strconv"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/service"
	"web-api/util"
)

var FbTransferTypeCalculateWager = map[string]int64{
	"WIN":                            -1,
	"CASHOUT":                        -1,
	"SETTLEMENT_ROLLBACK_RETURN":     1,
	"SETTLEMENT_ROLLBACK_DEDUCT":     1,
	"CASHOUT_CANCEL_DEDUCT":          1,
	"CASHOUT_CANCEL_RETURN":          1,
	"CASHOUT_CANCEL_ROLLBACK_DEDUCT": -1,
	"CASHOUT_CANCEL_ROLLBACK_RETURN": -1,
}

type Callback struct {
	Request     callback.OrderPayRequest
	Transaction ploutos.FbTransactionC
}

func (c *Callback) NewCallback(userId int64) {
	copier.Copy(&c.Transaction, &c.Request)
	c.Transaction.UserId = userId
	c.Transaction.Amount = int64(c.Request.Amount * 100)
}

func (c *Callback) GetGameVendorId() int64 {
	return consts.GameVendor["fb"]
}

func (c *Callback) GetGameTransactionId() int64 {
	return c.Transaction.ID
}

func (c *Callback) GetExternalUserId() string {
	return c.Request.MerchantUserId
}

func (c *Callback) SaveGameTransaction(tx *gorm.DB) error {
	return tx.Save(&c.Transaction).Error
}

func (c *Callback) ShouldProceed() bool {
	return c.Transaction.Status != 0 // skip transactions with status 0
}

func (c *Callback) GetAmount() int64 {
	return c.Transaction.Amount
}

func (c *Callback) GetWagerMultiplier() (value int64, exists bool) {
	value, exists = FbTransferTypeCalculateWager[c.Transaction.TransferType]
	return
}

func (c *Callback) GetBetAmount() (amount int64, exists bool) {
	e := model.DB.Model(ploutos.FbTransactionC{}).Select(`amount`).
		Where(`business_id`, c.Transaction.BusinessId).
		Where(`transfer_type`, `BET`).Order(`id`).First(&amount).Error
	if e == nil {
		exists = true
	}
	return
}

func (c *Callback) IsAdjustment() bool {
	return false
}

const (
	MONGODB_FB_CALLBACK_SYNC_ORDERS  = "fb_callback_sync_orders"
	MONGODB_FB_CALLBACK_SYNC_CASHOUT = "fb_callback_sync_cashout"
)

func BalanceCallback(c *gin.Context, req callback.BalanceRequest) (res callback.BaseResponse, err error) {
	gpu, balance, _, _, err := service.GetUserAndSum(consts.GameVendor["fb"], req.MerchantUserId)
	if err != nil {
		return
	}
	currency, err := strconv.Atoi(gpu.ExternalCurrency)
	if err != nil {
		return
	}
	data := callback.BalanceResponse{
		Balance:    fmt.Sprintf("%.2f", float64(balance)/100),
		CurrencyId: int64(currency),
	}
	res = callback.BaseResponse{
		Code: 0,
		Data: []callback.BalanceResponse{data},
	}
	return
}

func OrderPayCallback(c *gin.Context, req callback.OrderPayRequest) (res callback.BaseResponse, err error) {
	j, _ := json.Marshal(req)
	fmt.Println("order_pay: ", string(j))
	err = service.ProcessTransaction(&Callback{Request: req})
	if err != nil {
		return
	}
	res = callback.BaseResponse{
		Code: 0,
	}
	return
}

func CheckOrderPayCallback(c *gin.Context, req callback.OrderPayRequest) (res callback.BaseResponse, err error) {
	j, _ := json.Marshal(req)
	fmt.Println("check_order_pay: ", string(j))
	var fbTx ploutos.FbTransaction
	rows := model.DB.Where(`transaction_id`, req.TransactionId).First(&fbTx).RowsAffected
	if rows == 1 {
		res = callback.BaseResponse{
			Code: 0,
		}
		return
	}
	return
}

func SyncTransactionCallback(c *gin.Context, req []callback.OrderPayRequest) (res callback.BaseResponse, err error) {
	j, _ := json.Marshal(req)
	fmt.Println("sync_transaction: ", string(j))
	go func(c *gin.Context, req []callback.OrderPayRequest) {
		for _, r := range req {
			jj, _ := json.Marshal(r)
			_, e := cache.RedisSyncTransactionClient.Set(context.TODO(), fmt.Sprintf(`fb:%s:%s`, r.MerchantUserId, r.TransactionId), jj, 0).Result()
			if e != nil {
				util.Log().Error("redis error", e)
			}
		}
	}(c, req)
	res = callback.BaseResponse{
		Code: 0,
	}
	return
}

func SyncOrdersCallback(c *gin.Context, req callback.SyncOrdersRequest) (res callback.BaseResponse, err error) {
	j, _ := json.Marshal(req)
	fmt.Println("sync_orders: ", string(j))
	go func(c *gin.Context, req callback.SyncOrdersRequest) {
		coll := model.MongoDB.Collection(MONGODB_FB_CALLBACK_SYNC_ORDERS)
		_, e := coll.InsertOne(context.TODO(), req)
		if e != nil {
			util.Log().Error("mongodb error", e)
		}
	}(c, req)
	res = callback.BaseResponse{
		Code: 0,
	}
	return
}

func SyncCashoutCallback(c *gin.Context, req callback.SyncCashoutRequest) (res callback.BaseResponse, err error) {
	j, _ := json.Marshal(req)
	fmt.Println("sync_cashout: ", string(j))
	go func(c *gin.Context, req callback.SyncCashoutRequest) {
		coll := model.MongoDB.Collection(MONGODB_FB_CALLBACK_SYNC_CASHOUT)
		_, e := coll.InsertOne(context.TODO(), req)
		if e != nil {
			util.Log().Error("mongodb error", e)
		}
	}(c, req)
	res = callback.BaseResponse{
		Code: 0,
	}
	return
}
