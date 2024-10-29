package cash_method_promotion

import (
	"context"
	"log"

	"web-api/model"
	"web-api/service/common"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"

	"github.com/google/uuid"
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
		ctx = rfcontext.AppendDescription(ctx, "orderCashMethodId == 0")
		log.Printf(rfcontext.Fmt(ctx))
		return
	}
	{
		// check claim before or not
		promotionRecord, err := PromotionRecordByCashOrderId(orderId, nil)
		if err != nil {
			ctx = rfcontext.AppendError(ctx, err, "orderCashMethodId == 0")
			log.Printf(rfcontext.Fmt(ctx))
			return
		}

		if promotionRecord.ID != 0 {
			ctx = rfcontext.AppendDescription(ctx, "promotionRecord.ID != 0. exist -> skip")
			log.Printf(rfcontext.Fmt(ctx))
			return
		}
	}

	vipRecord, err := model.GetVipWithDefault(ctx, orderUserId)
	vipRecordVipRuleId := vipRecord.VipRule.ID

	ctx = rfcontext.AppendParams(ctx, callDesc, map[string]any{
		"vipRecordVipRuleId": vipRecordVipRuleId,
	})
	if err != nil {
		ctx = rfcontext.AppendError(ctx, err, "GetVipWithDefault")
		log.Printf(rfcontext.Fmt(ctx))
		return
	}

	ctx = rfcontext.AppendDescription(ctx, "GetVipWithDefault ok")
	log.Printf(rfcontext.Fmt(ctx))

	// check cash method and vip combination has promotion or not
	cashMethodPromotion, err := ByCashMethodIdAndVipId(model.DB, orderCashMethodId, vipRecordVipRuleId, &cashedInOrder.CreatedAt, &cashedInOrder.AppliedCashInAmount)
	cashMethodPromotionId := cashMethodPromotion.ID
	ctx = rfcontext.AppendParams(ctx, callDesc, map[string]any{
		"cashMethodPromotionId": cashMethodPromotionId,
	})

	if err != nil {
		ctx = rfcontext.AppendError(ctx, err, "ByCashMethodIdAndVipId")
		log.Printf(rfcontext.Fmt(ctx))
		return
	}
	if cashMethodPromotionId == 0 {
		ctx = rfcontext.AppendDescription(ctx, "cashMethodPromotionId == 0")
		log.Printf(rfcontext.Fmt(ctx))
		return
	}

	ctx = rfcontext.AppendDescription(ctx, "Get promo ok")
	log.Printf(rfcontext.Fmt(ctx))

	// check over payout limit or not
	claimedPast7DaysL, claimedPast1DayL, err := GetAccumulatedClaimedCashMethodPromotionPast7And1Days(ctx, &orderCashMethodId, orderUserId)
	if err != nil {
		ctx = rfcontext.AppendError(ctx, err, "get user past claimed")
		log.Printf(rfcontext.Fmt(ctx))
		return
	}

	var claimedPast7Days int64
	if len(claimedPast7DaysL) > 0 {
		claimedPast7Days = claimedPast7DaysL[0].Amount // len(claimedPast7DaysL[0]) at most 1
	}
	var claimedPast1Day int64
	if len(claimedPast1DayL) > 0 {
		claimedPast1Day = claimedPast1DayL[0].Amount // len(claimedPast1DayL[0]) at most 1
	}

	cashOrderAmount := cashedInOrder.AppliedCashInAmount
	if cashedInOrder.OrderType == ploutos.CashOrderTypeCashOut {
		cashOrderAmount = cashedInOrder.AppliedCashOutAmount
	}

	// catch underflow
	if cashOrderAmount < cashMethodPromotion.MinPayout {
		ctx = rfcontext.AppendDescription(ctx, "cash order amount underflow")
		log.Printf(rfcontext.Fmt(ctx))
		return
	}

	ctx = rfcontext.AppendParams(ctx, callDesc, map[string]any{
		"claimedPast7Days": claimedPast7Days,
		"claimedPast1Day":  claimedPast1Day,
	})

	finalPayout, err := FinalPossiblePayout(ctx, claimedPast7Days, claimedPast1Day, cashMethodPromotion, cashOrderAmount, false)
	if err != nil {
		ctx = rfcontext.AppendError(ctx, err, "final payout cal err")
		log.Printf(rfcontext.Fmt(ctx))
		return
	}

	if finalPayout == 0 {
		ctx = rfcontext.AppendDescription(ctx, "final payout 0")
		log.Printf(rfcontext.Fmt(ctx))
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
		updatedSum, err := model.UpdateDbUserSumAndCreateTransaction(rfcontext.AppendCallDesc(rfcontext.Spawn(context.Background()), "ValidateAndClaim"), tx, orderUserId, finalPayout, wagerChange, 0, ploutos.TransactionTypeCashMethodPromotion, "")
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
			ctx = rfcontext.AppendError(ctx, err, "tx.Create(&dummyOrder).Error")
			log.Printf(rfcontext.Fmt(ctx))
			return err
		}
		go common.SendUserSumSocketMsg(orderUserId, updatedSum.UserSum, "promotion", float64(finalPayout)/100)
		return
	})
	if err != nil {
		ctx = rfcontext.AppendError(ctx, err, "transaction")
		log.Printf(rfcontext.Fmt(ctx))
		return
	}
}
