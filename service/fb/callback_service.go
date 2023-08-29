package fb

import (
	"blgit.rfdev.tech/taya/game-service/fb/callback"
	models "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/util"
)

const MONGODB_FB_CALLBACK_SYNC_TRANSACTION = "fb_callback_sync_transaction"

func BalanceCallback(c *gin.Context, req callback.BalanceRequest) (res callback.BaseResponse, err error) {
	gpu, err := GetGameProviderUser(consts.GameProvider["fb"], req.MerchantUserId)
	if err != nil {
		res = callbackErrorResponse(c, req, err)
		return
	}

	balance, _, _, err := GetSums(gpu)
	if err != nil {
		res = callbackErrorResponse(c, req, err)
		return
	}

	data := callback.BalanceResponse{
		Balance: fmt.Sprintf("%.2f", float64(balance) / 100),
		CurrencyId: gpu.ExternalCurrencyId,
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
	gpu, err := GetGameProviderUser(consts.GameProvider["fb"], req.MerchantUserId)
	if err != nil {
		res = callbackErrorResponse(c, req, err)
		return
	}

	balance, remainingWager, maxWithdrawable, err := GetSums(gpu)
	if err != nil {
		res = callbackErrorResponse(c, req, err)
		return
	}

	err = ProcessTransaction(req, gpu.UserId, balance, remainingWager, maxWithdrawable)
	if err != nil {
		res = callbackErrorResponse(c, req, err)
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
	var fbTx model.FbTransaction
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
			coll := model.MongoDB.Collection(MONGODB_FB_CALLBACK_SYNC_TRANSACTION)
			_, e := coll.InsertOne(context.TODO(), r)
			if e != nil {
				util.Log().Error("mongodb error", e)
			}
		}
	}(c, req)
	go func(c *gin.Context, req []callback.OrderPayRequest) {
		for _, r := range req {
			jj, _ := json.Marshal(r)
			_, e := cache.RedisSyncTransactionClient.Set(context.TODO(), fmt.Sprintf(`%s:%s`, r.MerchantUserId, r.TransactionId), jj, 0).Result()
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

func ProcessTransaction(req callback.OrderPayRequest, userId int64, balance int64, remainingWager int64, maxWithdrawable int64) (err error) {
	fbTx := model.FbTransaction{
		models.FbTransactionC{
			UserId: userId,
			TransactionId: req.TransactionId,
			ExternalUserId: req.UserId,
			MerchantId: req.MerchantId,
			MerchantUserId: req.MerchantUserId,
			BusinessId: req.BusinessId,
			TransactionType: req.TransactionType,
			TransferType: req.TransferType,
			ExternalCurrencyId: req.CurrencyId,
			Amount: int64(req.Amount * 100),
			Status: req.Status,
			RelatedId: req.RelatedId,
		},
	}
	if fbTx.Status == 0 { // skip transactions with status 0
		return
	}

	tx := model.DB.Begin()
	if tx.Error != nil {
		err = tx.Error
		return
	}
	newBalance := balance + fbTx.Amount
	newRemainingWager := remainingWager
	if w, e := calWager(fbTx, remainingWager); e == nil {
		newRemainingWager = w
	}
	newWithdrawable := maxWithdrawable
	if w, e := calMaxWithdrawable(fbTx, newBalance, newRemainingWager, maxWithdrawable); e == nil {
		newWithdrawable = w
	}
	userSum := model.UserSum{
		models.UserSumC{
			Balance: newBalance,
			RemainingWager: newRemainingWager,
			MaxWithdrawable: newWithdrawable,
		},
	}
	rows := tx.Select(`balance`, `remaining_wager`, `max_withdrawable`).Where(`user_id`, userId).Updates(userSum).RowsAffected
	if rows == 0 {
		err = errors.New("insufficient balance or invalid transaction")
		tx.Rollback()
		return
	}
	err = tx.Save(&fbTx).Error
	if err != nil {
		tx.Rollback()
		return
	}
	transaction := model.Transaction{
		models.TransactionC{
			UserId:          userId,
			Amount:          fbTx.Amount,
			BalanceBefore:   balance,
			BalanceAfter:    userSum.Balance,
			FbTransactionId: fbTx.ID,
			Wager:           userSum.RemainingWager - remainingWager,
			WagerBefore:     remainingWager,
			WagerAfter:      userSum.RemainingWager,
		},
	}
	err = tx.Save(&transaction).Error
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	return
}

func calWager(fbTx model.FbTransaction, originalWager int64) (newWager int64, err error) {
	newWager = originalWager
	if !consts.FbTransferTypeCalculateWager[fbTx.TransferType] {
		return
	}
	var betTx model.FbTransaction
	err = model.DB.Where(`business_id`, fbTx.BusinessId).Where(`transfer_type`, `BET`).First(&betTx).Error
	if err != nil {
		return
	}
	absBetAmount := abs(betTx.Amount)
	wager := abs(absBetAmount - abs(fbTx.Amount))
	if wager > absBetAmount {
		wager = absBetAmount
	}
	newWager = originalWager - wager
	if newWager < 0 {
		newWager = 0
	}
	return
}

func calMaxWithdrawable(fbTx model.FbTransaction, balance int64, remainingWager int64, originalWithdrawable int64) (newWithdrawable int64, err error) {
	newWithdrawable = originalWithdrawable
	if !consts.FbTransferTypeCalculateWager[fbTx.TransferType] {
		return
	}
	if remainingWager == 0 {
		if balance > originalWithdrawable {
			newWithdrawable = balance
		}
	}
	return
}

func GetGameProviderUser(provider int64, userId string) (gpu model.GameProviderUser, err error) {
	err = gpu.GetByProviderAndExternalUser(provider, userId)
	return
}

func GetSums(gpu model.GameProviderUser) (balance int64, remainingWager int64, maxWithdrawable int64, err error) {
	var userSum model.UserSum
	err = model.DB.Where(`user_id`, gpu.UserId).First(&userSum).Error
	if err != nil {
		return
	}
	balance = userSum.Balance
	remainingWager = userSum.RemainingWager
	maxWithdrawable = userSum.MaxWithdrawable
	return
}

func callbackErrorResponse(c *gin.Context, req any, err error) (res callback.BaseResponse) {
	res = callback.BaseResponse{
		Code: 1,
		Message: err.Error(),
	}
	util.Log().Error(res.Message, c.Request.URL, c.Request.Header, util.MarshalService(req))
	return
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}