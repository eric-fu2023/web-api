package cash_method_promotion

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

// ValidateAndClaim
// validates applicable rules and net payout
// reward exist in table == claimed
func ValidateAndClaim(ctx context.Context, cashedInOrder model.CashOrder) {
	callDesc := "ValidateAndClaim"

	orderId := cashedInOrder.ID
	orderUserId := cashedInOrder.UserId
	orderCashMethodId := cashedInOrder.CashMethodId

	ctx = rfcontext.AppendParams(ctx, callDesc, map[string]any{
		"orderId":           orderId,
		"orderUserId":       orderUserId,
		"orderCashMethodId": orderCashMethodId,
	})

	if orderCashMethodId == 0 {
		util.GetLoggerEntry(ctx).Info("ValidateAndClaim orderCashMethodId == 0", orderId)
		return
	}
	{
		// check claim before or not
		promotionRecord, err := PromotionRecordByCashOrderId(orderId, nil)
		if err != nil {
			util.GetLoggerEntry(ctx).Error("ValidateAndClaim PromotionRecordByCashOrderId", err, orderId)
			return
		}

		if promotionRecord.ID != 0 {
			util.GetLoggerEntry(ctx).Info("ValidateAndClaim PromotionRecordByCashOrderId order has been claimed", orderId)
			return
		}
	}

	vipRecord, err := model.GetVipWithDefault(ctx, orderUserId)
	vipRecordVipRuleId := vipRecord.VipRule.ID

	ctx = rfcontext.AppendParams(ctx, callDesc, map[string]any{
		"vipRecordVipRuleId": vipRecordVipRuleId,
	})
	if err != nil {
		util.GetLoggerEntry(ctx).Error(rfcontext.Fmt(rfcontext.AppendDescription(ctx, fmt.Sprintf("ValidateAndClaim GetVipWithDefault %v %v", err, orderId))))
		return
	}

	util.GetLoggerEntry(ctx).Info("ValidateAndClaim vipRecordVipRuleId", vipRecordVipRuleId, orderId) // wl: for staging debug

	// check cash method and vip combination has promotion or not
	cashMethodPromotion, err := PromoByCashMethodIdAndVipId(orderCashMethodId, vipRecordVipRuleId, &cashedInOrder.CreatedAt, &cashedInOrder.AppliedCashInAmount, model.DB)
	cashMethodPromotionId := cashMethodPromotion.ID
	ctx = rfcontext.AppendParams(ctx, callDesc, map[string]any{
		"cashMethodPromotionId": cashMethodPromotionId,
	})

	if err != nil {
		util.GetLoggerEntry(ctx).Error("ValidateAndClaim Find ValidateAndClaim", err, orderId)
		return
	}
	if cashMethodPromotionId == 0 {
		util.GetLoggerEntry(ctx).Error("ValidateAndClaim no ValidateAndClaim", orderId)
		return
	}

	util.GetLoggerEntry(ctx).Info("ValidateAndClaim cashMethodPromotionId", cashMethodPromotionId, orderId) // wl: for staging debug

	// check over payout limit or not
	claimedPast7DaysL, claimedPast1DayL, err := GetAccumulatedClaimedCashMethodPromotionPast7And1Days(ctx, orderCashMethodId, orderUserId)
	if err != nil {
		util.GetLoggerEntry(ctx).Error("ValidateAndClaim GetAccumulatedClaimedCashMethodPromotionPast7And1Days", err, orderId)
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

	cashOrderAmount := cashedInOrder.AppliedCashInAmount
	if cashedInOrder.OrderType == ploutos.CashOrderTypeCashOut {
		cashOrderAmount = cashedInOrder.AppliedCashOutAmount
	}

	// catch underflow
	if cashOrderAmount < cashMethodPromotion.MinPayout {
		util.GetLoggerEntry(ctx).Info(rfcontext.Fmt(rfcontext.AppendDescription(ctx, fmt.Sprintf("Juicyy - MinPayout bigger than dep/wd cashOrderAmount, so no promotion. orderId: %s ", orderId)))) // staging debug
		return
	}

	finalPayout, err := FinalPayout(ctx, weeklyAmount, dailyAmount, cashMethodPromotion, cashOrderAmount, false)
	if err != nil {
		util.GetLoggerEntry(ctx).Error("ValidateAndClaim GetMaxAmountPayment", err, orderId)
		return
	}

	if finalPayout == 0 {
		util.GetLoggerEntry(ctx).Info("ValidateAndClaim finalPayout == 0", orderId)
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
		util.GetLoggerEntry(ctx).Info("ValidateAndClaim new cash order ", cashOrderId, orderId) // wl: for staging debug
		common.SendUserSumSocketMsg(orderUserId, updatedSum.UserSum, "promotion", float64(finalPayout)/100)
		return
	})
	if err != nil {
		ctx = rfcontext.AppendError(ctx, err, "ValidateAndClaim")
		log.Printf(rfcontext.Fmt(ctx))
		util.GetLoggerEntry(ctx).Error("ValidateAndClaim", err, orderId)
		return
	}
}
