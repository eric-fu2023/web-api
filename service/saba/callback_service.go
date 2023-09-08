package saba

import (
	"blgit.rfdev.tech/taya/game-service/saba/callback"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/service"
)

func GetBalanceCallback(c *gin.Context, req callback.GetBalanceRequest) (res any, err error) {
	gpu, err := service.GetGameProviderUser(consts.GameProvider["saba"], req.Message.UserId)
	if err != nil {
		return
	}

	balance, _, _, err := service.GetSums(gpu)
	if err != nil {
		return
	}

	now := time.Now().In(time.FixedZone("GMT-4", -4*60*60))
	res = callback.GetBalanceResponse{
		BaseResponse: callback.BaseResponse{
			Status: "0",
		},
		UserId:    req.Message.UserId,
		Balance:   float64(balance) / 100,
		BalanceTs: now.Format(time.RFC3339),
	}
	return
}

func PlaceBetCallback(c *gin.Context, req callback.PlaceBetRequest) (res any, err error) {
	j, _ := json.Marshal(req)
	fmt.Println("placebet: ", string(j))
	gpu, err := service.GetGameProviderUser(consts.GameProvider["saba"], req.Message.UserId)
	if err != nil {
		return
	}

	balance, wager, _, err := service.GetSums(gpu)
	if err != nil {
		return
	}

	var sabaTx model.SabaTransaction
	copier.Copy(&sabaTx, &req.Message)
	if v, e := time.Parse(time.RFC3339, req.Message.KickOffTime); e == nil {
		sabaTx.KickOffTime = v.UTC()
	}
	if v, e := time.Parse(time.RFC3339, req.Message.BetTime); e == nil {
		sabaTx.BetTime = v.UTC()
	}
	if v, e := time.Parse(time.RFC3339, req.Message.UpdateTime); e == nil {
		sabaTx.UpdateTime = v.UTC()
	}
	if v, e := time.Parse(time.RFC3339, req.Message.MatchDatetime); e == nil {
		sabaTx.MatchDatetime = v.UTC()
	}
	sabaTx.UserId = gpu.UserId
	sabaTx.ExternalUserId = req.Message.UserId
	sabaTx.BetAmount = int64(req.Message.BetAmount * 100)
	sabaTx.ActualAmount = int64(req.Message.ActualAmount * 100)
	sabaTx.CreditAmount = int64(req.Message.CreditAmount * 100)
	sabaTx.DebitAmount = int64(req.Message.DebitAmount * 100)

	tx := model.DB.Begin()
	if tx.Error != nil {
		err = tx.Error
		return
	}
	amount := -1 * sabaTx.ActualAmount
	userSum := model.UserSum{
		Balance: balance + amount,
	}
	rows := tx.Select(`balance`).Where(`user_id`, gpu.UserId).Updates(userSum).RowsAffected
	if rows == 0 {
		err = errors.New("insufficient balance or invalid transaction")
		tx.Rollback()
		return
	}
	err = tx.Save(&sabaTx).Error
	if err != nil {
		tx.Rollback()
		return
	}
	transaction := model.Transaction{
		UserId:            gpu.UserId,
		Amount:            amount,
		BalanceBefore:     balance,
		BalanceAfter:      userSum.Balance,
		SabaTransactionId: sabaTx.ID,
		Wager:             0,
		WagerBefore:       wager,
		WagerAfter:        wager,
	}
	err = tx.Save(&transaction).Error
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()

	res = callback.PlaceBetResponse{
		BaseResponse: callback.BaseResponse{
			Status: "0",
		},
		RefId:        req.Message.RefId,
		LicenseeTxId: sabaTx.ID,
	}
	return
}

func ConfirmBetCallback(c *gin.Context, req callback.ConfirmBetRequest) (res any, err error) {
	j, _ := json.Marshal(req)
	fmt.Println("confirmbet: ", string(j))
	gpu, err := service.GetGameProviderUser(consts.GameProvider["saba"], req.Message.UserId)
	if err != nil {
		return
	}

	balance, wager, _, err := service.GetSums(gpu)
	if err != nil {
		return
	}

	var sabaTx model.SabaTransaction
	var changedAmount int64
	for _, txn := range req.Message.Txns {
		rows := model.DB.Where(`ref_id`, txn.RefId).First(&sabaTx).RowsAffected
		if rows == 0 {
			continue
		}
		sabaTx.CfmOperationId = req.Message.OperationId
		if v, e := time.Parse(time.RFC3339, req.Message.UpdateTime); e == nil {
			sabaTx.CfmUpdateTime = v.UTC()
		}
		if v, e := time.Parse(time.RFC3339, req.Message.TransactionTime); e == nil {
			sabaTx.CfmTransactionTime = v.UTC()
		}
		sabaTx.CfmTxId = txn.TxId
		sabaTx.CfmIsOddsChanged = txn.IsOddsChanged
		if v, e := time.Parse(time.RFC3339, txn.WinlostDate); e == nil {
			sabaTx.CfmWinlostDate = v.UTC()
		}
		sabaTx.ActualAmount = int64(txn.ActualAmount * 100)
		changedAmount = int64(txn.CreditAmount * 100)
		sabaTx.DebitAmount = sabaTx.DebitAmount - changedAmount // confirmbet only refunds money to the user from more favourable odds
	}
	tx := model.DB.Begin()
	if tx.Error != nil {
		err = tx.Error
		return
	}
	newBalance := balance + changedAmount
	if changedAmount != 0 {
		rows := tx.Model(model.UserSum{}).Where(`user_id`, gpu.UserId).Update("balance", gorm.Expr("balance + ?", changedAmount)).RowsAffected
		if rows == 0 {
			err = errors.New("invalid transaction")
			tx.Rollback()
			return
		}
		transaction := model.Transaction{
			UserId:            gpu.UserId,
			Amount:            changedAmount,
			BalanceBefore:     balance,
			BalanceAfter:      newBalance,
			SabaTransactionId: sabaTx.ID,
			Wager:             0,
			WagerBefore:       wager,
			WagerAfter:        wager,
			IsAdjustment:      true,
		}
		err = tx.Save(&transaction).Error
		if err != nil {
			tx.Rollback()
			return
		}
	}
	err = tx.Save(&sabaTx).Error
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()

	res = callback.ConfirmCancelBetResponse{
		BaseResponse: callback.BaseResponse{
			Status: "0",
		},
		Balance: float64(newBalance) / 100,
	}
	return
}

func CancelBetCallback(c *gin.Context, req callback.CancelBetRequest) (res any, err error) {
	j, _ := json.Marshal(req)
	fmt.Println("cancelbet: ", string(j))
	gpu, err := service.GetGameProviderUser(consts.GameProvider["saba"], req.Message.UserId)
	if err != nil {
		return
	}

	balance, wager, _, err := service.GetSums(gpu)
	if err != nil {
		return
	}

	newBalance := balance
	for _, txn := range req.Message.Txns {
		var sabaTx model.SabaTransaction
		var changedAmount int64

		if rows := model.DB.Where(`ref_id`, txn.RefId).First(&sabaTx).RowsAffected; rows == 0 {
			continue
		}
		if sabaTx.CancOperationId != "" { // if succeeded before, skip that cancelbet txn
			continue
		}
		sabaTx.CancOperationId = req.Message.OperationId
		if v, e := time.Parse(time.RFC3339, req.Message.UpdateTime); e == nil {
			t := v.UTC()
			sabaTx.CancUpdateTime = &t
		}
		changedAmount = int64(txn.CreditAmount * 100)
		sabaTx.ActualAmount = sabaTx.ActualAmount - changedAmount
		sabaTx.DebitAmount = sabaTx.DebitAmount - changedAmount

		tx := model.DB.Begin()
		if tx.Error != nil {
			err = tx.Error
			return
		}
		newBalance += changedAmount
		if rows := tx.Model(model.UserSum{}).Where(`user_id`, gpu.UserId).Update("balance", gorm.Expr("balance + ?", changedAmount)).RowsAffected; rows == 0 {
			err = errors.New("invalid transaction")
			tx.Rollback()
			return
		}
		transaction := model.Transaction{
			UserId:            gpu.UserId,
			Amount:            changedAmount,
			BalanceBefore:     balance,
			BalanceAfter:      newBalance,
			SabaTransactionId: sabaTx.ID,
			Wager:             0,
			WagerBefore:       wager,
			WagerAfter:        wager,
			IsAdjustment:      true,
		}
		err = tx.Save(&transaction).Error
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Save(&sabaTx).Error
		if err != nil {
			tx.Rollback()
			return
		}
		tx.Commit()
	}

	res = callback.ConfirmCancelBetResponse{
		BaseResponse: callback.BaseResponse{
			Status: "0",
		},
		Balance: float64(newBalance) / 100,
	}
	return
}
