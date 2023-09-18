package cashout

import (
	"web-api/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CloseCashOutOrder(c *gin.Context, orderNumber string, actualAmount, bonusAmount, additionalWagerChange int64, notes, account, remark string, txDB *gorm.DB) (updatedCashOrder model.CashOrder, err error) {
	var newCashOrderState model.CashOrder
	err = txDB.Transaction(func(tx *gorm.DB) (err error) {
		err = txDB.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id", orderNumber).First(&newCashOrderState).Error
		if err != nil {
			return
		}
		newCashOrderState.ActualCashOutAmount = actualAmount
		newCashOrderState.BonusCashOutAmount = bonusAmount
		newCashOrderState.EffectiveCashOutAmount = newCashOrderState.AppliedCashOutAmount + bonusAmount
		newCashOrderState.Notes = notes
		newCashOrderState.WagerChange += additionalWagerChange
		newCashOrderState.Account = account
		newCashOrderState.Remark = remark
		newCashOrderState.Status = 2
		// update cash order
		err = tx.Where("id", orderNumber).Updates(newCashOrderState).Error
		if err != nil {
			return
		}
		updatedCashOrder = newCashOrderState
		return
	})

	return
}
