package cashin

import (
	"context"

	"web-api/conf/consts"
	"web-api/model"
	"web-api/service/common"
	"web-api/service/social_media_pixel"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
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
func CloseCashInOrder(c *gin.Context, orderNumber string, actualAmount, bonusAmount, additionalWagerChange int64, notes string, txDB *gorm.DB, transactionType int64) (updatedCashOrder model.CashOrder, err error) {
	var newCashOrderState model.CashOrder
	err = txDB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
		newCashOrderState, err = model.CashOrder{}.GetPendingOrPeApWithLockWithDB(orderNumber, tx)
		if err != nil {
			return
		}
		newCashOrderState.ActualCashInAmount = actualAmount
		newCashOrderState.BonusCashInAmount = bonusAmount
		newCashOrderState.EffectiveCashInAmount = newCashOrderState.AppliedCashInAmount + bonusAmount
		newCashOrderState.Notes = ploutos.EncryptedStr(notes)
		newCashOrderState.WagerChange += additionalWagerChange
		newCashOrderState.Status = ploutos.CashOrderStatusSuccess
		updatedCashOrder, err = closeOrder(newCashOrderState, tx, transactionType)
		if err != nil {
			return
		}
		// only if the cash orders has been settled, then we start adding the rewards
		// err = cash_orders.CreateReferralRewardRecord(txDB, newCashOrderState.UserId, orderNumber)
		// if err != nil {
		// 	util.GetLoggerEntry(c).Error("CreateReferralRewardRecord error", err)
		// 	return
		// }
		return
	})

	return
}

func closeOrder(newCashOrderState model.CashOrder, txDB *gorm.DB, transactionType int64) (updatedCashOrder model.CashOrder, err error) {
	// update cash order
	err = txDB.Omit(clause.Associations).Updates(newCashOrderState).Error
	// modify user sum
	if err != nil {
		return
	}
	userSum, err := model.UpdateDbUserSumAndCreateTransaction(txDB,
		newCashOrderState.UserId,
		newCashOrderState.EffectiveCashInAmount,
		newCashOrderState.WagerChange,
		0,
		transactionType,
		newCashOrderState.ID)
	if err != nil {
		return
	}

	updatedCashOrder = newCashOrderState

	common.SendCashNotificationWithoutCurrencyId(newCashOrderState.UserId, consts.Notification_Type_Cash_Transaction, common.NOTIFICATION_DEPOSIT_SUCCESS_TITLE, common.NOTIFICATION_DEPOSIT_SUCCESS, newCashOrderState.AppliedCashInAmount)
	common.SendUserSumSocketMsg(newCashOrderState.UserId, userSum.UserSum, "deposit_success", float64(updatedCashOrder.AppliedCashInAmount)/100)
	return
}

// deprecated: FE will do the reporting instead
func HandleSmPixelReporting(c context.Context, order model.CashOrder) {
	// Get user
	var user model.User
	if err := model.DB.Where(`id`, order.UserId).First(&user).Error; err != nil {
		util.GetLoggerEntry(c).Error("get user error", err)
		return
	}
	paymentDetails := social_media_pixel.PaymentDetails{
		Currency:     "USD",
		Value:        order.AppliedCashInAmount,
		CashMethodId: order.CashMethodId,
	}

	social_media_pixel.ReportPayment(c, user, paymentDetails)
}
