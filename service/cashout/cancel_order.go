package cashout

import (
	"errors"
	"web-api/model"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// newStatus = 3, 5, 7
func RevertCashOutOrder(c *gin.Context, orderNumber string, notes, remark string, newStatus int64, txDB *gorm.DB) (updatedCashOrder model.CashOrder, err error) {
	var newCashOrderState model.CashOrder
	switch newStatus {
	case models.CashOrderStatusCancelled, models.CashOrderStatusRejected, models.CashOrderStatusFailed:
	default:
		err = errors.New("wrong status")
		return
	}
	err = txDB.Debug().WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id", orderNumber).
			Where("status in ?", append(models.CashOrderStatusCollectionNonTerminal), models.CashOrderStatusFailed).
			First(&newCashOrderState).Error
		if err != nil || newCashOrderState.Status == models.CashOrderStatusFailed {
			return
		}
		newCashOrderState.Notes = notes
		newCashOrderState.Remark += remark
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
