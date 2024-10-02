package common

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"web-api/model"
	"web-api/util"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/plugin/dbresolver"
)

// ProcessImUpdateBalanceTransaction imsb update balance callbacks and allows negative user sum balance
// see ProcessImUpdateBalanceTransactionWithoutWagerCalculation for the version without wager calc.
// oct 2 2024 wager calculation migrated to use bet report.
func ProcessImUpdateBalanceTransaction(ctx context.Context, transactionRequest CallbackInterface) (err error) {
	ctx = rfcontext.AppendCallDesc(ctx, "ProcessImUpdateBalanceTransaction")

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
	fmt.Printf("DebugLog1234: GameVendorId=%d, newBalance=%d\n", transactionRequest.GetGameVendorId(), newBalance)
	reducedRemainingWager := remainingWager
	reducedDepositRemainingWager := depositRemainingWager
	//betAmount, betExists, w, depositWager, e := calculateWagerBatace(transactionRequest, remainingWager, reducedDepositRemainingWager)

	// FIXME
	// [x] calculateWagerBatace doesnt calculate wager correctly, need a new method calculateWagerBatace IMSB
	// [ ] deprecate this turnover calc, to derive turnover from bet report
	betAmount, betExists, w, depositWager, wagerChange, depositWagerChange, e := calculateWagerImsb(transactionRequest, remainingWager)
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

// ProcessImUpdateBalanceTransaction imsb update balance callbacks and allows negative user sum balance
// oct 2 2024 wager calculation migrated to use bet report.
func ProcessImUpdateBalanceTransactionWithoutWagerCalculation(ctx context.Context, transactionRequest CallbackInterface) (err error) {
	ctx = rfcontext.AppendCallDesc(ctx, "ProcessImUpdateBalanceTransactionWithoutWagerCalculation")

	tx := model.DB.Clauses(dbresolver.Use("txConn")).Begin(&sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if tx.Error != nil {
		err = tx.Error
		return
	}
	gpu, balance, _ /*remainingWager*/, _ /* depositRemainingWager */, maxWithdrawable, err := GetUserAndSumBatace(tx, transactionRequest.GetGameVendorId(), transactionRequest.GetExternalUserId())
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
	//reducedRemainingWager := remainingWager
	//reducedDepositRemainingWager := depositRemainingWager
	betAmount := transactionRequest.GetBetAmountOnly()
	betAmount, betExists, _ /*wager*/, _ /* depositWager*/, _ /*wagerChange*/, _ /*depositWagerChange*/, e := calculateWagerBatace(transactionRequest, 0, 0)
	//_, _ = wagerChange, depositWagerChange // not used. for wager data in `transaction` table, wager changes to update as before minus after
	//if e == nil {
	//	reducedRemainingWager = w
	//	reducedDepositRemainingWager = depositWager
	//}

	newWithdrawable := maxWithdrawable
	if w, e := calMaxWithdrawable(transactionRequest, newBalance, 0 /*reducedRemainingWager*/, maxWithdrawable); e == nil {
		newWithdrawable = w
	}
	userSum := ploutos.UserSum{
		Balance: newBalance,
		//RemainingWager:        reducedRemainingWager,
		//DepositRemainingWager: reducedDepositRemainingWager,
		MaxWithdrawable: newWithdrawable,
	}
	rows := tx.Select(`balance`, `remaining_wager`, `max_withdrawable`, `deposit_remaining_wager`).Where(`user_id`, gpu.UserId).Updates(userSum).RowsAffected
	if rows == 0 {
		err = ErrInsuffientBalance
		tx.Rollback()
		return
	}
	err = transactionRequest.SaveGameTransaction(tx)
	if err != nil {
		ctx = rfcontext.AppendError(ctx, err, "save game transactionRequest")
		log.Println(rfcontext.Fmt(ctx))
		tx.Rollback()
		return
	}
	transaction := ploutos.Transaction{
		UserId:               gpu.UserId,
		Amount:               transactionRequest.GetAmount(),
		BalanceBefore:        balance,
		BalanceAfter:         newBalance,
		ForeignTransactionId: transactionRequest.GetGameTransactionId(),
		//Wager:                userSum.RemainingWager - remainingWager,
		//WagerBefore:          remainingWager,
		//WagerAfter:           userSum.RemainingWager,
		IsAdjustment: transactionRequest.IsAdjustment(),
		//DepositWager:         userSum.DepositRemainingWager - reducedDepositRemainingWager,
		//DepositWagerBefore:   depositRemainingWager,
		//DepositWagerAfter:    userSum.DepositRemainingWager,
		GameVendorId: transactionRequest.GetGameVendorId(),
	}
	err = tx.Debug().Save(&transaction).Error
	if err != nil {
		ctx = rfcontext.AppendError(ctx, err, "save transaction")
		log.Println(rfcontext.Fmt(ctx))
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
