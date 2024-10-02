package common

import (
	"context"
	"database/sql"
	"fmt"

	"web-api/model"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/plugin/dbresolver"
)

// ProcessImUpdateBalanceTransaction imsb update balance callbacks and allows negative user sum balance
func ProcessImUpdateBalanceTransaction(ctx context.Context, transactionRequest CallbackInterface) (err error) {
	tx := model.DB.Clauses(dbresolver.Use("txConn")).Begin(&sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if tx.Error != nil {
		err = tx.Error
		return
	}
	gpu, balance, remainingWager, depositRemainingWager, maxWithdrawable, err := GetUserAndSumBatace(tx, transactionRequest.GetGameVendorId(), transactionRequest.GetExternalUserId())
	if err != nil {
		err = fmt.Errorf("%w: %w", ErrGameVendorUserInvalid, err)
		tx.Rollback()
		return
	}
	transactionRequest.NewCallback(gpu.UserId)
	if !transactionRequest.ShouldProceed() {
		tx.Rollback()
		return
	}
	newBalance := balance + transactionRequest.GetAmount()
	reducedRemainingWager := remainingWager
	reducedDepositRemainingWager := depositRemainingWager
	//betAmount, betExists, w, depositWager, e := calculateWagerBatace(transactionRequest, remainingWager, reducedDepositRemainingWager)

	betAmount, betExists, w, depositWager, wagerChange, depositWagerChange, e := calculateWagerBatace(transactionRequest, remainingWager, depositRemainingWager)
	_, _ = wagerChange, depositWagerChange // not used. for wager data in `transaction` table, wager changes to update as before minus after
	if e == nil {
		reducedRemainingWager = w
		reducedDepositRemainingWager = depositWager
	}

	newWithdrawable := maxWithdrawable
	if w, e := calMaxWithdrawable(transactionRequest, newBalance, reducedRemainingWager, maxWithdrawable); e == nil {
		newWithdrawable = w
	}
	userSum := ploutos.UserSum{
		Balance:               newBalance,
		RemainingWager:        reducedRemainingWager,
		DepositRemainingWager: reducedDepositRemainingWager,
		MaxWithdrawable:       newWithdrawable,
	}
	rows := tx.Select(`balance`, `remaining_wager`, `max_withdrawable`, `deposit_remaining_wager`).Where(`user_id`, gpu.UserId).Updates(userSum).RowsAffected
	if rows == 0 {
		err = ErrInsuffientBalance
		tx.Rollback()
		return
	}
	err = transactionRequest.SaveGameTransaction(tx)
	if err != nil {
		tx.Rollback()
		return
	}
	transaction := ploutos.Transaction{
		UserId:               gpu.UserId,
		Amount:               transactionRequest.GetAmount(),
		BalanceBefore:        balance,
		BalanceAfter:         newBalance,
		ForeignTransactionId: transactionRequest.GetGameTransactionId(),
		Wager:                userSum.RemainingWager - remainingWager,
		WagerBefore:          remainingWager,
		WagerAfter:           userSum.RemainingWager,
		IsAdjustment:         transactionRequest.IsAdjustment(),
		DepositWager:         userSum.DepositRemainingWager - reducedDepositRemainingWager,
		DepositWagerBefore:   depositRemainingWager,
		DepositWagerAfter:    userSum.DepositRemainingWager,
		GameVendorId:         transactionRequest.GetGameVendorId(),
	}
	err = tx.Save(&transaction).Error
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()

	e = transactionRequest.ApplyInsuranceVoucher(gpu.UserId, abs(betAmount), betExists)
	if e != nil {
		util.Log().Error("apply insurance voucher error: ", e.Error())
	}

	//SendNotification(gpu.UserId, consts.Notification_Type_Bet_Placement, NOTIFICATION_PLACE_BET_TITLE, NOTIFICATION_PLACE_BET)
	SendUserSumSocketMsg(gpu.UserId, userSum, "bet", float64(transactionRequest.GetAmount())/100)
	return
}
