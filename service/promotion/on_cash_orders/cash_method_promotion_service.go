package on_cash_orders

import (
	"context"
	"fmt"
	"log"

	"web-api/model"
	"web-api/service/common"
	"web-api/util"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// ValidateAndClaimCashMethodPromotion
// validates applicable rules and net payout
// reward exist in table == claimed
func ValidateAndClaimCashMethodPromotion(ctx context.Context, order model.CashOrder) {
	callDesc := "ValidateAndClaimCashMethodPromotion"

	orderId := order.ID
	orderUserId := order.UserId
	orderCashMethodId := order.CashMethodId

	ctx = rfcontext.AppendParams(ctx, callDesc, map[string]any{
		"orderId":           orderId,
		"orderUserId":       orderUserId,
		"orderCashMethodId": orderCashMethodId,
	})

	if orderCashMethodId == 0 {
		util.GetLoggerEntry(ctx).Info("ValidateAndClaimCashMethodPromotion orderCashMethodId == 0", orderId)
		return
	}
	{
		// check claim before or not
		cashMethodPromotionRecord, err := model.FindCashMethodPromotionRecordByCashOrderId(orderId, nil)
		if err != nil {
			util.GetLoggerEntry(ctx).Error("ValidateAndClaimCashMethodPromotion FindCashMethodPromotionRecordByCashOrderId", err, orderId)
			return
		}

		if cashMethodPromotionRecord.ID != 0 {
			util.GetLoggerEntry(ctx).Info("ValidateAndClaimCashMethodPromotion FindCashMethodPromotionRecordByCashOrderId order has been claimed", orderId)
			return
		}
	}

	vipRecord, err := model.GetVipWithDefault(ctx, orderUserId)
	vipRecordVipRuleId := vipRecord.VipRule.ID

	ctx = rfcontext.AppendParams(ctx, callDesc, map[string]any{
		"vipRecordVipRuleId": vipRecordVipRuleId,
	})
	if err != nil {
		util.GetLoggerEntry(ctx).Error(rfcontext.Fmt(rfcontext.AppendDescription(ctx, fmt.Sprintf("ValidateAndClaimCashMethodPromotion GetVipWithDefault %v %v", err, orderId))))
		return
	}

	util.GetLoggerEntry(ctx).Info("ValidateAndClaimCashMethodPromotion vipRecordVipRuleId", vipRecordVipRuleId, orderId) // wl: for staging debug

	// check cash method and vip combination has promotion or not
	cashMethodPromotion, err := model.FindActiveCashMethodPromotionByCashMethodIdAndVipId(orderCashMethodId, vipRecordVipRuleId, &order.CreatedAt, &order.AppliedCashInAmount, model.DB)
	cashMethodPromotionId := cashMethodPromotion.ID
	ctx = rfcontext.AppendParams(ctx, callDesc, map[string]any{
		"cashMethodPromotionId": cashMethodPromotionId,
	})

	if err != nil {
		util.GetLoggerEntry(ctx).Error("ValidateAndClaimCashMethodPromotion Find ValidateAndClaimCashMethodPromotion", err, orderId)
		return
	}
	if cashMethodPromotionId == 0 {
		util.GetLoggerEntry(ctx).Error("ValidateAndClaimCashMethodPromotion no ValidateAndClaimCashMethodPromotion", orderId)
		return
	}

	util.GetLoggerEntry(ctx).Info("ValidateAndClaimCashMethodPromotion cashMethodPromotionId", cashMethodPromotionId, orderId) // wl: for staging debug

	// check over payout limit or not
	claimedPast7DaysL, claimedPast1DayL, err := model.GetAccumulatedClaimedCashMethodPromotionPast7And1Days(ctx, orderCashMethodId, orderUserId)
	if err != nil {
		util.GetLoggerEntry(ctx).Error("ValidateAndClaimCashMethodPromotion GetAccumulatedClaimedCashMethodPromotionPast7And1Days", err, orderId)
		return
	}

	var weeklyAmount int64
	if len(claimedPast7DaysL) > 0 {
		weeklyAmount = claimedPast7DaysL[0].Amount // len(claimedPast7DaysL[0]) at most 1
	}
	var dailyAmount int64
	if len(claimedPast1DayL) > 0 {
		dailyAmount = claimedPast1DayL[0].Amount // len(claimedPast1DayL[0]) at most 1
	}

	cashOrderAmount := order.AppliedCashInAmount
	if order.OrderType == ploutos.CashOrderTypeCashOut {
		cashOrderAmount = order.AppliedCashOutAmount
	}

	// catch underflow
	if cashOrderAmount < cashMethodPromotion.MinPayout {
		util.GetLoggerEntry(ctx).Info(rfcontext.Fmt(rfcontext.AppendDescription(ctx, fmt.Sprintf("Juicyy - MinPayout bigger than dep/wd cashOrderAmount, so no promotion. orderId: %s ", orderId)))) // staging debug
		return
	}

	finalPayout, err := model.GetCashMethodPromotionFinalPayout(ctx, weeklyAmount, dailyAmount, cashMethodPromotion, cashOrderAmount, false)
	if err != nil {
		util.GetLoggerEntry(ctx).Error("ValidateAndClaimCashMethodPromotion GetMaxAmountPayment", err, orderId)
		return
	}

	if finalPayout == 0 {
		util.GetLoggerEntry(ctx).Info("ValidateAndClaimCashMethodPromotion finalPayout == 0", orderId)
		return
	}

	// claim start
	err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(ctx).Transaction(func(tx *gorm.DB) (err error) {
		newCashMethodPromotionRecord := ploutos.CashMethodPromotionRecord{
			CashMethodPromotionId: cashMethodPromotionId,
			CashMethodId:          cashMethodPromotion.CashMethodId,
			VipId:                 vipRecordVipRuleId,
			UserId:                orderUserId,
			PayoutRate:            cashMethodPromotion.PayoutRate,
			CashOrderId:           orderId,
			Amount:                finalPayout,
		}
		err = tx.Create(&newCashMethodPromotionRecord).Error
		if err != nil {
			return err
		}
		wagerChange := finalPayout
		updatedSum, err := model.UpdateDbUserSumAndCreateTransaction(tx, orderUserId, finalPayout, wagerChange, 0, ploutos.TransactionTypeCashMethodPromotion, "")
		if err != nil {
			return err
		}
		notes := "dummy" // is dummy important or not???

		cashOrderId := uuid.NewString()
		dummyOrder := ploutos.CashOrder{
			ID:                    cashOrderId,
			UserId:                orderUserId,
			OrderType:             ploutos.CashOrderTypeCashMethodPromotion,
			Status:                ploutos.CashOrderStatusSuccess,
			Notes:                 ploutos.EncryptedStr(notes),
			AppliedCashInAmount:   finalPayout,
			ActualCashInAmount:    finalPayout,
			EffectiveCashInAmount: finalPayout,
			BalanceBefore:         updatedSum.Balance - finalPayout,
			WagerChange:           finalPayout,
		}
		err = tx.Create(&dummyOrder).Error
		if err != nil {
			return err
		}
		util.GetLoggerEntry(ctx).Info("ValidateAndClaimCashMethodPromotion new cash order ", cashOrderId, orderId) // wl: for staging debug
		common.SendUserSumSocketMsg(orderUserId, updatedSum.UserSum, "promotion", float64(finalPayout)/100)
		return
	})
	if err != nil {
		ctx = rfcontext.AppendError(ctx, err, "ValidateAndClaimCashMethodPromotion")
		log.Printf(rfcontext.Fmt(ctx))
		util.GetLoggerEntry(ctx).Error("ValidateAndClaimCashMethodPromotion", err, orderId)
		return
	}
}
