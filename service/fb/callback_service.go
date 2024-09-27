package fb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/service/common"
	"web-api/service/promotion"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/fb/callback"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var FbTransferTypeCalculateWager = map[string]int64{
	"BET":                            0, // not for wager calculation but for whether to proceed max_withdrawal calculation
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
	Transaction ploutos.FbTransaction
}

func (c *Callback) NewCallback(userId int64) {
	copier.Copy(&c.Transaction, &c.Request)
	c.Transaction.ExternalUserId = c.Request.UserId
	c.Transaction.ExternalCurrencyId = c.Request.CurrencyId
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

func (c *Callback) GetBetAmountOnly() int64 {
	return 0
}
func (c *Callback) GetWagerMultiplier() (value int64, exists bool) {
	var txn ploutos.FbTransaction
	rows := model.DB.Clauses(dbresolver.Use("txConn")). // for re-settle without rollback
								Where(`business_id`, c.Transaction.BusinessId).
								Where(`amount`, 0).
								Where(`transfer_type`, `WIN`).Order(`id`).Find(&txn).RowsAffected
	if rows > 0 {
		return
	}
	value, exists = FbTransferTypeCalculateWager[c.Transaction.TransferType]
	return
}

func (c *Callback) GetBetAmount() (amount int64, exists bool) {
	e := model.DB.Clauses(dbresolver.Use("txConn")).Model(ploutos.FbTransaction{}).Select(`amount`).
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

func (c *Callback) ApplyInsuranceVoucher(userId int64, betAmount int64, betExists bool) (err error) {
	if c.Transaction.TransferType != "WIN" || !betExists || betAmount <= c.Transaction.Amount {
		return
	}

	var fbTx ploutos.FbTransaction
	err = model.DB.Clauses(dbresolver.Use("txConn")).Where(`business_id`, c.Transaction.BusinessId).Where(`transfer_type`, `BET`).
		Order(`id`).First(&fbTx).Error
	if err != nil {
		return
	}
	if fbTx.RelatedId == "" {
		return
	}

	err = model.DB.Clauses(dbresolver.Use("txConn")).Transaction(func(tx *gorm.DB) (err error) {
		voucherId, err := strconv.ParseInt(c.Transaction.RelatedId, 10, 64)
		if err != nil {
			return
		}
		ctx := context.TODO()
		voucher, err := model.VoucherPendingGetByIDUserWithDBNoTime(ctx, userId, voucherId, tx)
		if err != nil {
			return
		}

		var order callback.SyncOrdersRequest
		coll := model.MongoDB.Collection("fb_callback_sync_orders")
		filter := bson.M{"id": c.Transaction.BusinessId}
		opts := options.Find()
		opts.SetLimit(1)
		opts.SetSort(bson.D{{"createdAt", -1}})
		cursor, err := coll.Find(ctx, filter, opts)
		for cursor.Next(ctx) {
			cursor.Decode(&order)
		}
		if len(order.BetList) != 1 { // single (not parlay) should only have one match in betList
			return
		}
		floatOdd, err := strconv.ParseFloat(order.BetList[0].Odds, 64)
		if err != nil {
			return
		}

		if !promotion.ValidateVoucherUsageByType(voucher, int(order.BetList[0].OddsFormat), promotion.MatchTypeNotStarted, floatOdd, betAmount, false) {
			err = promotion.ErrVoucherUseInvalid
			return
		}

		err = tx.Model(&ploutos.Voucher{}).Where("id", voucher.ID).Updates(map[string]any{
			"status": ploutos.VoucherStatusRedeemed,
		}).Error
		if err != nil {
			return
		}

		rewardAmount := voucher.Amount
		loss := betAmount - c.Transaction.Amount
		if loss < rewardAmount {
			rewardAmount = loss
		}
		wagerChange := voucher.WagerMultiplier * rewardAmount
		err = promotion.CreateCashOrder(tx, voucher.PromotionType, userId, rewardAmount, wagerChange, "", "")
		if err != nil {
			return
		}
		return
	})
	return
}

const (
	MONGODB_FB_CALLBACK_SYNC_ORDERS  = "fb_callback_sync_orders"
	MONGODB_FB_CALLBACK_SYNC_CASHOUT = "fb_callback_sync_cashout"
)

func BalanceCallback(c *gin.Context, req callback.BalanceRequest) (res callback.BaseResponse, err error) {
	log.Printf("CallbackBalance fb.BalanceCallback() ... %#v \n", req)
	gpu, balance, _, _, err := common.GetUserAndSum(model.DB, consts.GameVendor["fb"], req.MerchantUserId)

	if err != nil {
		log.Printf("CallbackBalance fb.BalanceCallback() err %#v gpu %v, balance %v \n", err, gpu, balance)
		return
	}
	currency, err := strconv.Atoi(gpu.ExternalCurrency)
	if err != nil {
		log.Printf("CallbackBalance fb.BalanceCallback() ExternalCurrency err %#v gpu %v, balance %v \n", err, gpu, balance)
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
	log.Printf("CallbackBalance fb.BalanceCallback() END currency err %#v data %v, res %v \n", err, data, res)
	return
}

func OrderPayCallback(c *gin.Context, req callback.OrderPayRequest) (res callback.BaseResponse, err error) {
	go common.LogGameCallbackRequest("order_pay", req)
	err = common.ProcessTransaction(&Callback{Request: req})
	if err != nil {
		return
	}
	res = callback.BaseResponse{
		Code: 0,
	}
	return
}

func CheckOrderPayCallback(c *gin.Context, req callback.OrderPayRequest) (res callback.BaseResponse, err error) {
	go common.LogGameCallbackRequest("check_order_pay", req)
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
	go common.LogGameCallbackRequest("sync_transaction", req)
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
	go common.LogGameCallbackRequest("sync_orders", req)

	// insert into mongodb
	coll := model.MongoDB.Collection(MONGODB_FB_CALLBACK_SYNC_ORDERS)
	req.CreatedAt = time.Now().UnixMilli()
	_, e := coll.InsertOne(context.TODO(), req)
	if e != nil {
		util.Log().Error("mongodb error", e)
		res.Code = 1
		return
	}

	go func(c *gin.Context, req callback.SyncOrdersRequest) {
		if req.OrderStatus == 5 && req.SettleAmount == "0" {
			var transaction ploutos.FbTransaction
			e := model.DB.Where(`business_id`, req.Id).Where(`transfer_type = 'BET'`).Order(`created_at DESC`).First(&transaction).Error
			if e != nil {
				util.Log().Error("sync_orders settle amount 0 error", e)
				return
			}
			var obj ploutos.FbTransactionClone
			copier.Copy(&obj, &transaction)
			obj.TransactionId = fmt.Sprintf(`%d-%s`, req.Version, req.Id)
			obj.TransactionType = "IN"
			obj.TransferType = "WIN"
			obj.Amount = 0
			obj.Status = 1
			obj.RelatedId = req.RelatedId
			jj, e := json.Marshal(obj)
			if e != nil {
				util.Log().Error("sync_orders settle amount 0 error", e)
				return
			}
			_, e = cache.RedisSyncTransactionClient.Set(context.TODO(), fmt.Sprintf(`fb:%s:%s`, obj.MerchantUserId, obj.TransactionId), jj, 0).Result()
			if e != nil {
				util.Log().Error("sync_orders settle amount 0 error", e)
				return
			}
		}
	}(c, req)

	return
}

func SyncCashoutCallback(c *gin.Context, req callback.SyncCashoutRequest) (res callback.BaseResponse, err error) {
	go common.LogGameCallbackRequest("sync_cashout", req)
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
