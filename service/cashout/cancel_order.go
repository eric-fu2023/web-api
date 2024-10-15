package cashout

import (
	"context"
	"errors"

	"web-api/conf/consts"
	"web-api/model"
	"web-api/service/common"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
)

// newStatus = 3, 5, 7
func RevertCashOutOrder(c *gin.Context, orderNumber string, notes, remark string, newStatus int64, txDB *gorm.DB) (updatedCashOrder model.CashOrder, err error) {
	var newCashOrderState model.CashOrder
	switch newStatus {
	case ploutos.CashOrderStatusCancelled, ploutos.CashOrderStatusRejected, ploutos.CashOrderStatusFailed:
	default:
		err = errors.New("wrong status")
		return
	}
	err = txDB.Clauses(dbresolver.Use("txConn")).WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id", orderNumber).
			Where("status in ?", append(ploutos.CashOrderStatusCollectionNonTerminal, ploutos.CashOrderStatusFailed)).
			First(&newCashOrderState).Error
		if err != nil || newCashOrderState.Status == ploutos.CashOrderStatusFailed {
			return
		}
		newCashOrderState.Notes = ploutos.EncryptedStr(notes)
		newCashOrderState.Remark += remark
		newCashOrderState.Status = newStatus
		// update cash order
		err = tx.Omit(clause.Associations).Where("id", orderNumber).Updates(newCashOrderState).Error
		if err != nil {
			return
		}
		updatedCashOrder = newCashOrderState
		userSum, err := model.UpdateDbUserSumAndCreateTransaction(rfcontext.AppendCallDesc(rfcontext.Spawn(context.Background()), "revertCashOutOrder"),
			tx,
			newCashOrderState.UserId,
			newCashOrderState.AppliedCashOutAmount,
			newCashOrderState.WagerChange,
			newCashOrderState.AppliedCashOutAmount,
			10002,
			newCashOrderState.ID)

		common.SendCashNotificationWithoutCurrencyId(updatedCashOrder.UserId, consts.Notification_Type_Cash_Transaction, common.NOTIFICATION_WITHDRAWAL_FAILED_TITLE, common.NOTIFICATION_WITHDRAWAL_FAILED, updatedCashOrder.AppliedCashOutAmount)
		common.SendUserSumSocketMsg(newCashOrderState.UserId, userSum.UserSum, "withdraw_failed", float64(updatedCashOrder.AppliedCashOutAmount)/100)
		return
	})

	return
}
