package cashout

import (
	"web-api/conf/consts"
	"web-api/model"
	"web-api/service/common"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
)

func CloseCashOutOrder(c *gin.Context, orderNumber string, actualAmount, bonusAmount, additionalWagerChange int64, notes, remark string, allowPromotion bool, txDB *gorm.DB) (updatedCashOrder model.CashOrder, err error) {
	err = txDB.Clauses(dbresolver.Use("txConn")).WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id", orderNumber).
			Where("status in ?", append(ploutos.CashOrderStatusCollectionNonTerminal, ploutos.CashOrderStatusSuccess)).
			First(&updatedCashOrder).Error
		if err != nil || updatedCashOrder.Status == ploutos.CashOrderStatusSuccess {
			return
		}
		updatedCashOrder.ActualCashOutAmount = actualAmount
		updatedCashOrder.BonusCashOutAmount = bonusAmount
		updatedCashOrder.EffectiveCashOutAmount = updatedCashOrder.AppliedCashOutAmount + bonusAmount
		updatedCashOrder.Notes = ploutos.EncryptedStr(notes)
		updatedCashOrder.WagerChange += additionalWagerChange
		updatedCashOrder.Remark += remark
		updatedCashOrder.Status = ploutos.CashOrderStatusSuccess
		// update cash order
		err = tx.Where("id", orderNumber).Updates(updatedCashOrder).Error
		if err != nil {
			return
		}

		common.SendCashNotificationWithoutCurrencyId(updatedCashOrder.UserId, consts.Notification_Type_Cash_Transaction, common.NOTIFICATION_WITHDRAWAL_SUCCESS_TITLE, common.NOTIFICATION_WITHDRAWAL_SUCCESS, updatedCashOrder.AppliedCashOutAmount)
		go func() {
			userSum, _ := model.GetByUserIDWithLockWithDB(updatedCashOrder.UserId, model.DB)
			common.SendUserSumSocketMsg(updatedCashOrder.UserId, userSum.UserSum, "withdraw_success", float64(updatedCashOrder.AppliedCashOutAmount)/100)
		}()

		return
	})

	return
}
