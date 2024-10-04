package common

import (
	"database/sql"
	"fmt"

	"web-api/model"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
)

// ProcessTransactionBatace dollar jackpot, stream games, etc
func ProcessTransactionBatace(obj CallbackInterface) (err error) {
	tx := model.DB.Clauses(dbresolver.Use("txConn")).Begin(&sql.TxOptions{Isolation: sql.LevelSerializable})
	if tx.Error != nil {
		err = tx.Error
		return
	}

	gpu, balance, remainingWager, depositRemainingWager, maxWithdrawable, err := GetUserAndSumBatace(tx, obj.GetGameVendorId(), obj.GetExternalUserId())
	if err != nil {
		err = fmt.Errorf("%w: %w", ErrGameVendorUserInvalid, err)
		tx.Rollback()
		return
	}
	obj.NewCallback(gpu.UserId)
	if !obj.ShouldProceed() {
		tx.Rollback()
		return
	}
	newBalance := balance + obj.GetAmount()
	if newBalance < 0 {
		err = ErrInsuffientBalance
		tx.Rollback()
		return
	}

	newRemainingWager := remainingWager
	newDepositRemainingWager := depositRemainingWager
	_, _, w, dw, wc, dwc, e := calculateWagerBatace(obj, remainingWager, depositRemainingWager)
	if e == nil {
		newRemainingWager = w
		newDepositRemainingWager = dw
	}

	newWithdrawable := maxWithdrawable
	if w, e := calMaxWithdrawable(obj, newBalance, newRemainingWager, maxWithdrawable); e == nil {
		newWithdrawable = w
	}

	userSum := ploutos.UserSum{
		Balance:               newBalance,
		RemainingWager:        newRemainingWager,
		DepositRemainingWager: newDepositRemainingWager,
		MaxWithdrawable:       newWithdrawable,
	}

	rows := tx.Select(`balance`, `remaining_wager`, `deposit_remaining_wager`, `max_withdrawable`).Where(`user_id`, gpu.UserId).Updates(userSum).RowsAffected
	if rows == 0 {
		err = ErrInsuffientBalance
		tx.Rollback()
		return
	}
	err = obj.SaveGameTransaction(tx)
	if err != nil {
		tx.Rollback()
		return
	}
	transaction := ploutos.Transaction{
		UserId:               gpu.UserId,
		Amount:               obj.GetAmount(),
		BalanceBefore:        balance,
		BalanceAfter:         newBalance,
		ForeignTransactionId: obj.GetGameTransactionId(),
		Wager:                wc,
		WagerBefore:          remainingWager,
		WagerAfter:           userSum.RemainingWager,
		DepositWager:         dwc,
		DepositWagerBefore:   userSum.DepositRemainingWager,
		DepositWagerAfter:    newDepositRemainingWager,
		IsAdjustment:         obj.IsAdjustment(),
		GameVendorId:         obj.GetGameVendorId(),
	}
	err = tx.Save(&transaction).Error
	if err != nil {
		tx.Rollback()
		return
	}

	wagerAudit := ploutos.WagerAudit{
		SourceId:            string(obj.GetGameTransactionId()),
		UserId:              gpu.UserId,
		BeforeWager:         remainingWager,
		AfterWager:          userSum.RemainingWager,
		WagerChanges:        wc,
		DepositBeforeWager:  userSum.DepositRemainingWager,
		DepositAfterWager:   newDepositRemainingWager,
		DepositWagerChanges: dwc,
		SourceType:          ploutos.SourceTypeInternalGame,
	}

	err = tx.Create(&wagerAudit).Error
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()

	//SendNotification(gpu.UserId, consts.Notification_Type_Bet_Placement, NOTIFICATION_PLACE_BET_TITLE, NOTIFICATION_PLACE_BET)
	SendUserSumSocketMsg(gpu.UserId, userSum, "bet", float64(obj.GetAmount())/100)
	return
}

func GetUserAndSumBatace(tx *gorm.DB, gameVendorId int64, externalUserId string) (gameVendorUser ploutos.GameVendorUser, balance int64, remainingWager int64, depositRemainingWager int64, maxWithdrawable int64, err error) {
	gameVendorUser, err = GetGameVendorUser(gameVendorId, externalUserId)
	if err != nil {
		return
	}
	balance, remainingWager, depositRemainingWager, maxWithdrawable, err = GetSumsBatace(tx, gameVendorUser)
	if err != nil {
		return
	}
	return
}

func GetSumsBatace(tx *gorm.DB, gpu ploutos.GameVendorUser) (balance int64, remainingWager int64, depositRemainingWager int64, maxWithdrawable int64, err error) {
	var userSum ploutos.UserSum
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(`user_id`, gpu.UserId).First(&userSum).Error
	if err != nil {
		return
	}
	balance = userSum.Balance
	remainingWager = userSum.RemainingWager
	depositRemainingWager = userSum.DepositRemainingWager
	maxWithdrawable = userSum.MaxWithdrawable
	return
}

// calculateWagerBatace dollar jackpot, stream games, imsb,  ...
func calculateWagerBatace(transaction CallbackInterface, originalWager int64, originalDepositWager int64) (betAmount int64, betExists bool, newWager int64, newDepositWager int64, wagerChange int64, depositWagerChange int64, err error) {
	newWager = originalWager
	newDepositWager = originalDepositWager
	betAmount = transaction.GetBetAmountOnly()
	if !betExists {
		return
	}
	wagerChange = -betAmount
	newWager = newWager + wagerChange
	if newWager < 0 {
		newWager = 0
	}

	depositWagerChange = -betAmount
	newDepositWager = newDepositWager + depositWagerChange
	if newDepositWager < 0 {
		newDepositWager = 0
	}
	return
}

// calculateWagerBatace dollar jackpot, stream games, imsb,  ...
func calculateWagerImsb(transaction CallbackInterface, remainingTurnover int64) (betAmount int64, betExists bool, remainingTurnover2 int64, newDepositWager int64, wagerChange int64, depositWagerChange int64, err error) {
	remainingTurnover2 = remainingTurnover

	multiplier, exists := transaction.GetWagerMultiplier()
	betAmount, betExists = transaction.GetBetAmount()
	if !exists || !betExists {
		return
	}

	betAmount = abs(betAmount)
	turnoverToReduce := abs(betAmount - abs(transaction.GetAmount()))

	if turnoverToReduce > betAmount {
		turnoverToReduce = betAmount
	}
	remainingTurnover2 = remainingTurnover + (multiplier * turnoverToReduce)

	if remainingTurnover2 < 0 {
		remainingTurnover2 = 0
	}

	return
}
