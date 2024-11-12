package cashin

import (
	"context"
	"fmt"
	"log"
	"time"

	"web-api/conf/consts"
	"web-api/model"
	"web-api/service"
	"web-api/service/common"
	"web-api/service/promotion"
	"web-api/service/social_media_pixel"
	"web-api/util"

	"blgit.rfdev.tech/taya/common-function/cash_orders"
	"blgit.rfdev.tech/taya/common-function/rfcontext"
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
func CloseCashInOrder(c *gin.Context, ctx context.Context, orderNumber string, actualAmount, bonusAmount, additionalWagerChange int64, notes string, txDB *gorm.DB, transactionType int64) (updatedCashOrder model.CashOrder, err error) {
	ctx = rfcontext.AppendCallDesc(ctx, "CloseCashInOrder")
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
		updatedCashOrder, err = closeOrder(ctx, newCashOrderState, tx, transactionType)
		if err != nil {
			return
		}
		return
	})

	// if no error in closing the cash-in order, and the order must be auto cash in - (operation type 0) or make up order - (operation type ploutos.CashOrderOperationTypeMakeUpOrder)
	if err == nil && newCashOrderState.OrderType == ploutos.CashOrderTypeCashIn && (newCashOrderState.OperationType == ploutos.CashOrderOperationTypeMakeUpOrder || newCashOrderState.OperationType == 0) {
		uid := newCashOrderState.UserId
		ctx = rfcontext.AppendParams(ctx, "promotions", map[string]interface{}{
			"userId": uid,
		})
		now := time.Now().UTC()
		// this is to claim FTD bonus!!!
		ftdPromoCtx := rfcontext.AppendCallDesc(ctx, "this is to claim FTD bonus!!!")
		var ftdPromo ploutos.Promotion
		var ftdSession ploutos.PromotionSession
		pErr := txDB.Debug().Where("is_active").Where("type", ploutos.PromotionTypeFirstDepB).Where("start_at < ? and end_at > ?", now, now).First(&ftdPromo).Error
		if pErr != nil {
			ftdPromoCtx = rfcontext.AppendError(ftdPromoCtx, pErr, "promotion.First(&ftdPromo)")
		}
		ftdSession, pErr = model.GetActivePromotionSessionByPromotionId(ftdPromoCtx, ftdPromo.ID, now)
		if pErr != nil {
			ftdPromoCtx = rfcontext.AppendError(ftdPromoCtx, pErr, "promotion.GetActivePromotionSessionByPromotionId")
		}
		// if claim success, will send notification, and create notification in db.
		_, pErr = promotion.Claim(ftdPromoCtx, now, ftdPromo, ftdSession, uid, nil)
		if pErr != nil {
			ftdPromoCtx = rfcontext.AppendError(ftdPromoCtx, pErr, "promotion.Claim")
		}
		go log.Println(rfcontext.Fmt(rfcontext.AppendDescription(ftdPromoCtx, "END")))

		// this is to claim referral bonus!!!

		referralPromoCtx := rfcontext.AppendCallDesc(ctx, "this is to claim referral bonus!!!")
		var referralPromo ploutos.Promotion
		var referralSession ploutos.PromotionSession
		var userReferral ploutos.UserReferral
		pErr = txDB.Debug().Where("deleted_at is null").Where("referral_id = ?", uid).First(&userReferral).Error
		if pErr != nil {
			referralPromoCtx = rfcontext.AppendError(referralPromoCtx, pErr, "promotion.First(&userReferral)")
		}
		pErr = txDB.Debug().Where("is_active").Where("type", ploutos.PromotionTypeVipReferral).Where("start_at < ? and end_at > ?", now, now).First(&referralPromo).Error
		if pErr != nil {
			referralPromoCtx = rfcontext.AppendError(referralPromoCtx, pErr, "promotion.First(&referralPromo)")
		}
		referralSession, pErr = model.GetActivePromotionSessionByPromotionId(referralPromoCtx, referralPromo.ID, now)
		if pErr != nil {
			referralPromoCtx = rfcontext.AppendError(referralPromoCtx, pErr, "GetActivePromotionSessionByPromotionId")
		}
		// if claim success, will send notification, and create notification in db.
		_, pErr = promotion.Claim(referralPromoCtx, now, referralPromo, referralSession, userReferral.ReferrerId, nil)
		if pErr != nil {
			referralPromoCtx = rfcontext.AppendError(referralPromoCtx, pErr, "promotion.Claim")
		}
		go log.Println(rfcontext.Fmt(rfcontext.AppendDescription(referralPromoCtx, "END")))
	}
	return /* err */
}

func closeOrder(ctx context.Context, newCashOrderState model.CashOrder, txDB *gorm.DB, transactionType int64) (updatedCashOrder model.CashOrder, err error) {
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

	ctx = rfcontext.AppendDescription(ctx, "closeOrder ok")
	log.Println(rfcontext.Fmt(ctx))

	common.SendCashNotificationWithoutCurrencyId(newCashOrderState.UserId, consts.Notification_Type_Cash_Transaction, common.NOTIFICATION_DEPOSIT_SUCCESS_TITLE, common.NOTIFICATION_DEPOSIT_SUCCESS, newCashOrderState.AppliedCashInAmount)
	// only send notification when it is real deposit (make up order or real cash in)
	if newCashOrderState.OperationType == ploutos.CashOrderOperationTypeMakeUpOrder || newCashOrderState.OperationType == 0 {
		if is_FTD {
			common.SendUserSumSocketMsg(newCashOrderState.UserId, userSum.UserSum, "FTD_success", float64(updatedCashOrder.AppliedCashInAmount)/100)
			// if FTD success, need to send to pixel
			var user ploutos.User
			err = txDB.Debug().Table("users").Where("id = ?", userSum.UserId).Scan(&user).Error
			if err != nil {
				log.Printf("pixel app send data log error when finding user channel code")
			}
			if user.Channel == "pixel_app_001"{
				log.Printf("should log pixel event deposit for channel pixel_app_001")
				service.PixelFTDEvent(newCashOrderState.UserId, "0.0.0.0", newCashOrderState.AppliedCashInAmount)
			}
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
