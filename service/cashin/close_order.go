package cashin

import (
	"context"
	"fmt"
	"time"

	"web-api/conf/consts"
	"web-api/model"
	"web-api/service/common"
	"web-api/service/social_media_pixel"
	"web-api/util"

	"blgit.rfdev.tech/taya/common-function/cash_orders"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"blgit.rfdev.tech/taya/common-function/rfcontext"

	"web-api/service/promotion"

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
		return
	})

	// if no error in closing the cash-in order, and the order must be auto cash in - (operation type 0) or make up order - (operation type ploutos.CashOrderOperationTypeMakeUpOrder)
	if err == nil && newCashOrderState.OrderType == ploutos.CashOrderTypeCashIn && (newCashOrderState.OperationType == ploutos.CashOrderOperationTypeMakeUpOrder || newCashOrderState.OperationType == 0) {
		uid := newCashOrderState.UserId
		now := time.Now().UTC()
		// this is to claim FTD bonus!!!
		var ftdPromo ploutos.Promotion
		var ftdSession ploutos.PromotionSession
		err = txDB.Debug().Where("is_active").Where("type", ploutos.PromotionTypeFirstDepB).Where("start_at < ? and end_at > ?", now, now).First(&ftdPromo).Error
		if err != nil {
			fmt.Println("ftdPromo get err ", err)
		}
		ftdSession, err = model.GetActivePromotionSessionByPromotionId(context.TODO(), ftdPromo.ID, now)
		if err != nil {
			fmt.Println("ftdSession session get err ", err)
		}
		// if claim success, will send notification, and create notification in db.
		_, err = promotion.Claim(context.TODO(), now, ftdPromo, ftdSession, uid, nil)
		if err != nil {
			fmt.Println("ftdPromo.Claim err ", err)
		}
		fmt.Println("ftdPromo.Claim finished ", uid)


		// this is to claim referral bonus!!!
		var referralPromo ploutos.Promotion
		var referralSession ploutos.PromotionSession
		var userReferral ploutos.UserReferral
		err = txDB.Debug().Where("deleted_at is null").Where("referral_id = ?", uid).First(&userReferral).Error
		if err != nil {
			fmt.Println("user referral get err ", err)
		}
		err = txDB.Debug().Where("is_active").Where("type", ploutos.PromotionTypeVipReferral).Where("start_at < ? and end_at > ?", now, now).First(&referralPromo).Error
		if err != nil {
			fmt.Println("referralPromo get err ", err)
		}
		referralSession, err := model.GetActivePromotionSessionByPromotionId(context.TODO(), referralPromo.ID, now)
		if err != nil {
			fmt.Println("referralSession session get err ", err)
		}
		// if claim success, will send notification, and create notification in db.
		_, err = promotion.Claim(context.TODO(), now, referralPromo, referralSession, userReferral.ReferrerId, nil)
		if err != nil {
			fmt.Println("referralPromo.Claim err ", err)
		}
		fmt.Println("referralPromo.Claim finished ", uid)
	}

	return
}

func closeOrder(newCashOrderState model.CashOrder, txDB *gorm.DB, transactionType int64) (updatedCashOrder model.CashOrder, err error) {
	// update cash order
	err = txDB.Omit(clause.Associations).Updates(newCashOrderState).Error
	// modify user sum
	if err != nil {
		return
	}
	userSum, err := model.UpdateDbUserSumAndCreateTransaction(rfcontext.AppendCallDesc(rfcontext.Spawn(context.Background()), "CloseCashInOrder"),
		txDB,
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

	// only create for cash in order, for 2 types:
	// 0: real deposit
	// ploutos.CashOrderOperationTypeMakeUpOrder:  make up cash order
	if newCashOrderState.OrderType == ploutos.CashOrderTypeCashIn && (newCashOrderState.OperationType == ploutos.CashOrderOperationTypeMakeUpOrder || newCashOrderState.OperationType == 0) {
		err = cash_orders.CreateReferralRewardRecord(model.DB, newCashOrderState.ID)
		if err != nil {
			fmt.Println("CreateReferralRewardRecord for cash in order error", err)
			return
		}
	}

	// check if it is FTD
	is_FTD := false
	var cashOrder []model.CashOrder
	err = txDB.Where("user_id", userSum.UserId).
		Where("order_type", ploutos.CashOrderTypeCashIn).
		Where("status", ploutos.CashOrderStatusSuccess).
		Where("operation_type in (0, 5000)"). // 0 is for deposit from app, 5000 is for make up order
		Find(&cashOrder).Error
	if err != nil {
		return
	}
	// this order has been settled, so if this is the FTD, the length must be 1
	if len(cashOrder) == 1 {
		is_FTD = true
	}
	common.SendCashNotificationWithoutCurrencyId(newCashOrderState.UserId, consts.Notification_Type_Cash_Transaction, common.NOTIFICATION_DEPOSIT_SUCCESS_TITLE, common.NOTIFICATION_DEPOSIT_SUCCESS, newCashOrderState.AppliedCashInAmount)
	// only send notification when it is real deposit (make up order or real cash in)
	if newCashOrderState.OperationType == ploutos.CashOrderOperationTypeMakeUpOrder || newCashOrderState.OperationType == 0 {
		if is_FTD {
			common.SendUserSumSocketMsg(newCashOrderState.UserId, userSum.UserSum, "FTD_success", float64(updatedCashOrder.AppliedCashInAmount)/100)
		} else {
			common.SendUserSumSocketMsg(newCashOrderState.UserId, userSum.UserSum, "deposit_success", float64(updatedCashOrder.AppliedCashInAmount)/100)
		}
	} else {
		common.SendUserSumSocketMsg(newCashOrderState.UserId, userSum.UserSum, "other_cash_in_success", float64(updatedCashOrder.AppliedCashInAmount)/100)
	}

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
