package cashin

import (
	"context"
	"log"
	"strconv"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/service/common"
	"web-api/service/social_media_pixel"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
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
		newCashOrderState.Notes = models.EncryptedStr(notes)
		newCashOrderState.WagerChange += additionalWagerChange
		newCashOrderState.Status = 2
		updatedCashOrder, err = closeOrder(c, orderNumber, newCashOrderState, tx, transactionType)
		if err != nil {
			return
		}

		// 查看是否有砍单记录，添加进度到砍单任务
		go calculateTeamupSlashProgress(newCashOrderState.AppliedCashInAmount, newCashOrderState.UserId)
		return
	})
	if err == nil {
		go HandlePromotion(c.Copy(), newCashOrderState)
	}
	return
}

func closeOrder(c *gin.Context, orderNumber string, newCashOrderState model.CashOrder, txDB *gorm.DB, transactionType int64) (updatedCashOrder model.CashOrder, err error) {
	// update cash order

	err = txDB.Omit(clause.Associations).Updates(newCashOrderState).Error
	// modify user sum
	if err != nil {
		return
	}
	userSum, err := model.UserSum{}.UpdateUserSumWithDB(txDB,
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

func calculateTeamupSlashProgress(appliedCashInAmount, userId int64) {
	slashMultiplierString, err := model.GetAppConfigWithCache("teamup", "teamup_slash_multiplier")
	if err != nil {
		log.Printf("Get Slash Multiplier err - %v \n", err)
		return
	}
	slashMultiplier, err := strconv.Atoi(slashMultiplierString)
	if err != nil {
		log.Printf("Get Slash Multiplier err - %v \n", err)
		return
	}

	if appliedCashInAmount < int64(slashMultiplier) {
		return
	}

	// Convert cash amount into slash progress by dividing multiplier
	contributedSlashProgress := appliedCashInAmount / int64(slashMultiplier)
	err = model.GetTeamupProgressToUpdate(userId, appliedCashInAmount, contributedSlashProgress)
	if err != nil {
		log.Print(err.Error())
		return
	}
}
