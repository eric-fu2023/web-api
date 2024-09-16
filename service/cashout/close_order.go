package cashout

import (
	"web-api/conf/consts"
	"web-api/model"
	"web-api/service/common"
	"web-api/service/promotion/on_cash_orders"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
)

func CloseCashOutOrder(c *gin.Context, orderNumber string, actualAmount, bonusAmount, additionalWagerChange int64, notes, remark string, allowPromotion bool, txDB *gorm.DB, gateway on_cash_orders.PaymentGateway, requestMode on_cash_orders.RequestIngressMode) (updatedCashOrder model.CashOrder, err error) {
	err = txDB.Clauses(dbresolver.Use("txConn")).WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
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
		updatedCashOrder.Notes = models.EncryptedStr(notes)
		updatedCashOrder.WagerChange += additionalWagerChange
		updatedCashOrder.Remark += remark
		updatedCashOrder.Status = models.CashOrderStatusSuccess
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
	if err == nil {
		go func() {
			pErr := on_cash_orders.Handle(c, updatedCashOrder, models.TransactionTypeCashOut, on_cash_orders.CashOrderEventTypeClose, gateway, requestMode)
			if pErr != nil {
				util.Log().Error("cashin.CloseCashInOrder error on promotion handling", pErr)
			}
		}()
	}

	return
}
