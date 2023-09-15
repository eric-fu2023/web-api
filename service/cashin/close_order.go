package cashin

import (
	"web-api/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// check api response
// lock cash order
// update cash order
// {
// update user_sum
// create transaction history
// }

// modifiable in system call
// Status
// Notes
// ActualCashOutAmount
// BonusCashOutAmount
// EffectiveCashOutAmount
// ActualCashInAmount
// BonusCashInAmount
// EffectiveCashInAmount
// WagerChange -- only special cases
// Account
// Remark

func CloseCashInOrder(c *gin.Context, orderNumber string, actualAmount, bonusAmount, additionalWagerChange int64, notes, account, remark string, txDB *gorm.DB) (updatedCashOrder model.CashOrder, err error) {
	var newCashOrderState model.CashOrder
	err = txDB.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id", orderNumber).First(&newCashOrderState).Error
	if err != nil {
		return
	}
	newCashOrderState.ActualCashInAmount = actualAmount
	newCashOrderState.BonusCashInAmount = bonusAmount
	newCashOrderState.EffectiveCashInAmount = newCashOrderState.AppliedCashInAmount + bonusAmount
	newCashOrderState.Notes = notes
	newCashOrderState.WagerChange += additionalWagerChange
	newCashOrderState.Account = account
	newCashOrderState.Remark = remark
	newCashOrderState.Status = 2
	updatedCashOrder, err = closeOrder(c, orderNumber, newCashOrderState, txDB, 10000)
	return
}

func closeOrder(c *gin.Context, orderNumber string, newCashOrderState model.CashOrder, txDB *gorm.DB, transactionType int64) (updatedCashOrder model.CashOrder, err error) {
	err = txDB.Transaction(func(tx *gorm.DB) (err error) {
		// update cash order
		err = tx.Where("id", orderNumber).Updates(newCashOrderState).Error
		// modify user sum
		if err != nil {
			return
		}
		_, err = model.UserSum{}.UpdateUserSumWithDB(tx,
			newCashOrderState.UserId,
			newCashOrderState.EffectiveCashInAmount-newCashOrderState.EffectiveCashOutAmount,
			newCashOrderState.WagerChange,
			transactionType,
			newCashOrderState.ID)
		if err != nil {
			return
		}
		updatedCashOrder = newCashOrderState
		return
	})
	return
}
