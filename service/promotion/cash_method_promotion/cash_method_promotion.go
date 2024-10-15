package cash_method_promotion

import (
	"context"
	"time"

	"web-api/model"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

func PromoByCashMethodIdAndVipId(cashMethodId, vipId int64, promotionAt *time.Time, cashInAmount *int64, tx *gorm.DB) (cashMethodPromotion ploutos.CashMethodPromotion, err error) {
	if promotionAt == nil {
		now := time.Now().UTC()
		promotionAt = &now
	}

	if tx == nil {
		tx = model.DB
	}

	tx = tx.Debug().
		Where("cash_method_id", cashMethodId).Where("vip_id", vipId).
		Where("start_at < ? and end_at > ?", promotionAt, promotionAt).
		Where("status = ?", 1)

	// temporary guard for dev work, once stable can pass arg by value.
	if cashInAmount == nil {
		tx = tx.Where("? > floor_cash_in_amount", cashInAmount).Order("floor_cash_in_amount desc")
	}

	err = tx.First(&cashMethodPromotion).Error
	if err != nil {
		return
	}
	return
}

func FinalPayout(c context.Context, claimedPast7Days int64, claimedPast1Day int64, cashMethodPromotion ploutos.CashMethodPromotion, cashAmount int64, dryRun bool) (amount int64, err error) {
	if claimedPast7Days >= cashMethodPromotion.WeeklyMaxPayout {
		util.GetLoggerEntry(c).Info("FinalPayout claimedPast7Days >= cashMethodPromotion.WeeklyMaxPayout", claimedPast7Days, cashMethodPromotion.WeeklyMaxPayout)
		return
	}
	if claimedPast1Day >= cashMethodPromotion.DailyMaxPayout {
		util.GetLoggerEntry(c).Info("FinalPayout claimedPast1Day >= cashMethodPromotion.DailyMaxPayout", claimedPast1Day, cashMethodPromotion.DailyMaxPayout)
		return
	}

	ratedPayout := cashMethodPromotion.PayoutRate * float64(cashAmount)
	dailyClaimableCeiling := float64(cashMethodPromotion.DailyMaxPayout - claimedPast1Day)
	weeklyClaimableCeiling := float64(cashMethodPromotion.WeeklyMaxPayout - claimedPast7Days)

	if dryRun {
		return int64(min(dailyClaimableCeiling, weeklyClaimableCeiling)), nil
	} else {
		return int64(min(ratedPayout, dailyClaimableCeiling, weeklyClaimableCeiling)), nil
	}
}
