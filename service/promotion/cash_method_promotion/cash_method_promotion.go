package cash_method_promotion

import (
	"context"
	"errors"
	"time"

	"web-api/model"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

func ByCashMethodIdAndVipId(tx *gorm.DB, cashMethodId, vipId int64, promotionAt *time.Time, cashInAmount *int64) (cashMethodPromotion ploutos.CashMethodPromotion, err error) {
	if tx == nil {
		tx = model.DB
	}

	if promotionAt == nil {
		now := time.Now().UTC()
		promotionAt = &now
	}

	tx = tx.Debug().
		Where("cash_method_id", cashMethodId).Where("vip_id", vipId).
		Where("start_at < ? and end_at > ?", promotionAt, promotionAt).
		Where("status = ?", 1)

	// temporary guard for dev work, once stable can pass arg by value.
	if cashInAmount != nil {
		tx = tx.Where("? > floor_cash_in_amount", cashInAmount).Order("floor_cash_in_amount desc")
	} else {
		return cashMethodPromotion, errors.New("cashInAmount required")
	}

	err = tx.First(&cashMethodPromotion).Error
	return
}

// FinalPossiblePayout
// dryRun == calculate ceiling for the payout
func FinalPossiblePayout(c context.Context, claimedPast7Days int64, claimedPast1Day int64, cashMethodPromotion ploutos.CashMethodPromotion, cashAmount int64, dryRun bool) (amount int64, err error) {
	if claimedPast7Days >= cashMethodPromotion.WeeklyMaxPayout {
		util.GetLoggerEntry(c).Info("FinalPossiblePayout claimedPast7Days >= cashMethodPromotion.WeeklyMaxPayout", claimedPast7Days, cashMethodPromotion.WeeklyMaxPayout)
		return
	}
	if claimedPast1Day >= cashMethodPromotion.DailyMaxPayout {
		util.GetLoggerEntry(c).Info("FinalPossiblePayout claimedPast1Day >= cashMethodPromotion.DailyMaxPayout", claimedPast1Day, cashMethodPromotion.DailyMaxPayout)
		return
	}

	dailyClaimableCeiling := float64(cashMethodPromotion.DailyMaxPayout - claimedPast1Day)
	weeklyClaimableCeiling := float64(cashMethodPromotion.WeeklyMaxPayout - claimedPast7Days)

	if dryRun {
		return int64(min(dailyClaimableCeiling, weeklyClaimableCeiling)), nil
	} else {
		ratedPayout := cashMethodPromotion.PayoutRate * float64(cashAmount)
		return int64(min(ratedPayout, dailyClaimableCeiling, weeklyClaimableCeiling)), nil
	}
}
