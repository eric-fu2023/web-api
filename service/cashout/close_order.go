package cashout

import (
	"web-api/model"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CloseCashOutOrder(c *gin.Context, orderNumber string, actualAmount, bonusAmount, additionalWagerChange int64, notes, remark string, txDB *gorm.DB) (updatedCashOrder model.CashOrder, err error) {
	err = txDB.Debug().WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id", orderNumber).
			Where("status in ?", append(models.CashOrderStatusCollectionNonTerminal, models.CashOrderStatusSuccess)).
			First(&updatedCashOrder).Error
		if err != nil || updatedCashOrder.Status == models.CashOrderStatusSuccess {
			return
		}
		updatedCashOrder.ActualCashOutAmount = actualAmount
		updatedCashOrder.BonusCashOutAmount = bonusAmount
		updatedCashOrder.EffectiveCashOutAmount = updatedCashOrder.AppliedCashOutAmount + bonusAmount
		updatedCashOrder.Notes = notes
		updatedCashOrder.WagerChange += additionalWagerChange
		updatedCashOrder.Remark += remark
		updatedCashOrder.Status = 2
		// update cash order
		err = tx.Where("id", orderNumber).Updates(updatedCashOrder).Error
		if err != nil {
			return
		}
		return
	})

	return
}
