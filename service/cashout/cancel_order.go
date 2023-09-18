package cashout

import (
	"errors"
	"web-api/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// newStatus = 3, 5
func RevertCashOutOrder(c *gin.Context, orderNumber string, notes, account, remark string, newStatus int64, txDB *gorm.DB) (updatedCashOrder model.CashOrder, err error) {
	var newCashOrderState model.CashOrder
	switch newStatus {
	case 3, 5:
	default:
		err = errors.New("wrong status")
		return
	}
	err = txDB.Transaction(func(tx *gorm.DB) (err error) {
		err = txDB.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id", orderNumber).First(&newCashOrderState).Error
		if err != nil {
			return
		}

		newCashOrderState.Notes = notes
		newCashOrderState.Account = account
		newCashOrderState.Remark = remark
		newCashOrderState.Status = newStatus
		// update cash order
		err = tx.Where("id", orderNumber).Updates(newCashOrderState).Error
		if err != nil {
			return
		}
		updatedCashOrder = newCashOrderState
		_, err = model.UserSum{}.UpdateUserSumWithDB(
			tx,
			newCashOrderState.UserId,
			newCashOrderState.AppliedCashOutAmount,
			newCashOrderState.WagerChange,
			newCashOrderState.AppliedCashOutAmount,
			10002,
			newCashOrderState.ID)

		return
	})

	return
}
