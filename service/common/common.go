package common

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/json"
	"errors"
	"firebase.google.com/go/v4/messaging"
	"gorm.io/gorm"
	"web-api/model"
	"web-api/util"
)

type Platform struct {
	Platform int64 `form:"platform" json:"platform" binding:"required"`
}

type Page struct {
	Page  int `form:"page" json:"page" binding:"min=1"`
	Limit int `form:"limit" json:"limit" binding:"min=1"`
}

type UserRegisterInterface interface {
	CreateUser(model.User, string) error
	VendorRegisterError() error
	OthersError() error
}

type CallbackInterface interface {
	NewCallback(int64)
	GetGameVendorId() int64
	GetGameTransactionId() int64
	GetExternalUserId() string
	SaveGameTransaction(*gorm.DB) error
	ShouldProceed() bool
	GetAmount() int64
	GetWagerMultiplier() (int64, bool)
	GetBetAmount() (int64, bool)
	IsAdjustment() bool
}

func GetUserAndSum(gameVendor int64, externalUserId string) (gameVendorUser ploutos.GameVendorUser, balance int64, remainingWager int64, maxWithdrawable int64, err error) {
	gameVendorUser, err = GetGameVendorUser(gameVendor, externalUserId)
	if err != nil {
		return
	}
	balance, remainingWager, maxWithdrawable, err = GetSums(gameVendorUser)
	if err != nil {
		return
	}
	return
}

func GetGameVendorUser(vendor int64, userId string) (gpu ploutos.GameVendorUser, err error) {
	err = model.DB.Scopes(model.GameVendorUserByVendorAndExternalUser(vendor, userId)).First(&gpu).Error
	return
}

func GetSums(gpu ploutos.GameVendorUser) (balance int64, remainingWager int64, maxWithdrawable int64, err error) {
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
	gpu, balance, remainingWager, maxWithdrawable, err := GetUserAndSum(obj.GetGameVendorId(), obj.GetExternalUserId())
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
			UserId:               gpu.UserId,
			Amount:               obj.GetAmount(),
			BalanceBefore:        balance,
			BalanceAfter:         newBalance,
			ForeignTransactionId: obj.GetGameTransactionId(),
			TransactionType:      obj.GetGameVendorId(),
			Wager:                userSum.RemainingWager - remainingWager,
			WagerBefore:          remainingWager,
			WagerAfter:           userSum.RemainingWager,
			IsAdjustment:         obj.IsAdjustment(),
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
	betAmount, exists := obj.GetBetAmount()
	if !exists {
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

func SendNotification(user model.User, notificationType string, title string, text string) {
	go func() {
		notification := ploutos.NewUserNotification(user.UserC, text)
		if err := notification.Send(model.DB); err != nil {
			util.Log().Error("notification creation error: ", err.Error())
		}
	}()
	go func() {
		msgData := map[string]string{
			"notification_type": notificationType,
		}
		notification := messaging.Notification{
			Title: title,
			Body:  text,
		}
		tokens, err := model.GetFcmTokenStrings([]int64{user.ID})
		if err != nil {
			util.Log().Error("fcm token generation error: ", err.Error())
		}
		client := util.FCMFactory.NewClient(false)
		err = client.SendMessageToAll(msgData, notification, tokens)
		if err != nil {
			util.Log().Error("fcm sending error: ", err.Error())
		}
	}()
}

func LogGameCallbackRequest(action string, request any) {
	j, _ := json.Marshal(request)
	util.Log().Info(`%s: %s`, action, string(j))
}
