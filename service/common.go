package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"gorm.io/gorm"
	"web-api/model"
)

type Platform struct {
	Platform int64 `form:"platform" json:"platform" binding:"required"`
}

type Page struct {
	Page  int `form:"page" json:"page" binding:"min=1"`
	Limit int `form:"limit" json:"limit" binding:"min=1"`
}

type CallbackInterface interface {
	NewCallback(int64)
	GetGameProviderId() int64
	GetGameTransactionId() int64
	GetExternalUserId() string
	SaveGameTransaction(*gorm.DB) error
	ShouldProceed() bool
	GetAmount() int64
	GetWagerMultiplier() (int64, bool)
	GetBetAmount() (int64, error)
}

func GetUserAndSum(gameProvider int64, externalUserId string) (gameProviderUser ploutos.GameProviderUser, balance int64, remainingWager int64, maxWithdrawable int64, err error) {
	gameProviderUser, err = GetGameProviderUser(gameProvider, externalUserId)
	if err != nil {
		return
	}
	balance, remainingWager, maxWithdrawable, err = GetSums(gameProviderUser)
	if err != nil {
		return
	}
	return
}

func GetGameProviderUser(provider int64, userId string) (gpu ploutos.GameProviderUser, err error) {
	err = model.DB.Scopes(model.GameProviderUserByProviderAndExternalUser(provider, userId)).First(&gpu).Error
	return
}

func GetSums(gpu ploutos.GameProviderUser) (balance int64, remainingWager int64, maxWithdrawable int64, err error) {
	var userSum ploutos.UserSum
	err = model.DB.Where(`user_id`, gpu.UserId).First(&userSum).Error
	if err != nil {
		return
	}
	balance = userSum.Balance
	remainingWager = userSum.RemainingWager
	maxWithdrawable = userSum.MaxWithdrawable
	return
}

func ProcessTransaction(obj CallbackInterface) (err error) {
	gpu, balance, remainingWager, maxWithdrawable, err := GetUserAndSum(obj.GetGameProviderId(), obj.GetExternalUserId())
	if err != nil {
		return
	}
	obj.NewCallback(gpu.UserId)
	if !obj.ShouldProceed() {
		return
	}
	newBalance := balance + obj.GetAmount()
	newRemainingWager := remainingWager
	if w, e := calWager(obj, remainingWager); e == nil {
		newRemainingWager = w
	}
	newWithdrawable := maxWithdrawable
	if w, e := calMaxWithdrawable(obj, newBalance, newRemainingWager, maxWithdrawable); e == nil {
		newWithdrawable = w
	}
	userSum := ploutos.UserSum{
		ploutos.UserSumC{
			Balance:         newBalance,
			RemainingWager:  newRemainingWager,
			MaxWithdrawable: newWithdrawable,
		},
	}
	tx := model.DB.Begin()
	if tx.Error != nil {
		err = tx.Error
		return
	}
	rows := tx.Select(`balance`, `remaining_wager`, `max_withdrawable`).Where(`user_id`, gpu.UserId).Updates(userSum).RowsAffected
	if rows == 0 {
		err = errors.New("insufficient balance or invalid transaction")
		tx.Rollback()
		return
	}
	err = obj.SaveGameTransaction(tx)
	if err != nil {
		tx.Rollback()
		return
	}
	transaction := ploutos.Transaction{
		ploutos.TransactionC{
			UserId:            gpu.UserId,
			Amount:            obj.GetAmount(),
			BalanceBefore:     balance,
			BalanceAfter:      newBalance,
			ForeignTransactionId: obj.GetGameTransactionId(),
			Wager:             userSum.RemainingWager - remainingWager,
			WagerBefore:       remainingWager,
			WagerAfter:        userSum.RemainingWager,
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

func calWager(obj CallbackInterface, originalWager int64) (newWager int64, err error) {
	newWager = originalWager
	multiplier, exists := obj.GetWagerMultiplier()
	if !exists {
		return
	}
	betAmount, err := obj.GetBetAmount()
	if err != nil {
		return
	}
	absBetAmount := abs(betAmount)
	wager := abs(absBetAmount - abs(obj.GetAmount()))
	if wager > absBetAmount {
		wager = absBetAmount
	}
	newWager = originalWager + (multiplier * wager)
	if newWager < 0 {
		newWager = 0
	}
	return
}

func calMaxWithdrawable(obj CallbackInterface, balance int64, remainingWager int64, originalWithdrawable int64) (newWithdrawable int64, err error) {
	newWithdrawable = originalWithdrawable
	_, exists := obj.GetWagerMultiplier()
	if !exists {
		return
	}
	if remainingWager == 0 {
		if balance > originalWithdrawable {
			newWithdrawable = balance
		}
	}
	return
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
