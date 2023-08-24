package fb

import (
	"blgit.rfdev.tech/taya/game-service/fb/callback"
	models "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slices"
	"math"
	"web-api/conf/consts"
	"web-api/model"
)

func BalanceCallback(c *gin.Context, req callback.BalanceRequest) (res callback.BaseResponse, err error) {
	gpu, res := getGameProviderUser(consts.GameProvider["fb"], req.MerchantUserId)
	if res.Code != 0 {
		return
	}

	balance, _, res := getBalanceAndWager(gpu)
	if res.Code != 0 {
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
	gpu, res := getGameProviderUser(consts.GameProvider["fb"], req.MerchantUserId)
	if res.Code != 0 {
		return
	}

	balance, remainingWager, res := getBalanceAndWager(gpu)
	if res.Code != 0 {
		return
	}
	if balance < int64(req.Amount * 100) {
		res = callback.BaseResponse{
			Code: 9,
			Message: "insufficient balance",
		}
		return
	}

	err = processTransaction(c, req, gpu.UserId, balance, remainingWager)
	if err != nil {
		res = callback.BaseResponse{
			Code: 1,
			Message: err.Error(),
		}
		return
	}

	res = callback.BaseResponse{
		Code: 0,
	}
	return
}

func CheckOrderPayCallback(c *gin.Context, req callback.OrderPayRequest) (res callback.BaseResponse, err error) {
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

func SyncTransactionCallback(c *gin.Context, req callback.OrderPayRequest) (res callback.BaseResponse, err error) {

	return
}

func processTransaction(c *gin.Context, req callback.OrderPayRequest, userId int64, balance int64, remainingWager int64) (err error) {
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

	tx := model.DB.Begin()
	if tx.Error != nil {
		err = tx.Error
		return
	}
	absAmount := int64(math.Abs(float64(fbTx.Amount)))
	if slices.Contains(consts.TransferTypeNegativeWagerRequirement, fbTx.TransferType) {
		absAmount = -1 * absAmount
	}
	userSum := model.UserSum{
		models.UserSumC{
			Balance: balance + fbTx.Amount,
			RemainingWager: remainingWager + absAmount,
		},
	}
	rows := tx.Select(`balance`).Where(`user_id`, userId).Where(`balance`, balance).Updates(userSum).RowsAffected
	if rows == 0 {
		err = errors.New("duplicated or invalid transaction")
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
			UserId: userId,
			Amount: fbTx.Amount,
			BalanceBefore: balance,
			BalanceAfter: userSum.Balance,
			FbTransactionId: fbTx.ID,
			Wager: absAmount,
			WagerBefore: remainingWager,
			WagerAfter: userSum.RemainingWager,
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

func getGameProviderUser(provider int64, userId string) (gpu model.GameProviderUser, res callback.BaseResponse) {
	err := gpu.GetByProviderAndExternalUser(provider, userId)
	if err != nil {
		res = callback.BaseResponse{
			Code: 1,
			Message: err.Error(),
		}
		return
	}
	return
}

func getBalanceAndWager(gpu model.GameProviderUser) (balance int64, remainingWager int64, res callback.BaseResponse) {
	var userSum model.UserSum
	err := model.DB.Where(`user_id`, gpu.UserId).First(&userSum).Error
	if err != nil {
		res = callback.BaseResponse{
			Code: 1,
			Message: err.Error(),
		}
		return
	}
	balance = userSum.Balance
	remainingWager = userSum.RemainingWager
	return
}