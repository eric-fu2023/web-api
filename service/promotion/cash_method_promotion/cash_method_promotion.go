package cash_method_promotion

import (
	"context"
	"errors"
	"log"
	"time"

	"web-api/model"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
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
		tx = tx.Where("? >= min_payout", cashInAmount).Order("min_payout desc")
	} else {
		return cashMethodPromotion, errors.New("cashInAmount required")
	}

	err = tx.First(&cashMethodPromotion).Error
	return
}

type ConfigStat struct {
	PayoutRate_Max      ploutos.Fractional `json:"payout_rate_max"`
	DailyMaxPayout_Max  int64              `json:"daily_max_payout_max"`
	WeeklyMaxPayout_Max int64              `json:"weekly_max_payout_max"`
	MinPayout_Min       int64              `json:"min_payout_min"`
}

func ConfigStats(tx *gorm.DB, cashMethodId *int64, vipId *int64, promotionAt *time.Time) (ConfigStat, error) {
	if cashMethodId == nil || vipId == nil {
		return ConfigStat{}, errors.New("cashMethodId and vipId required to cal max payout rate")
	}
	if tx == nil {
		tx = model.DB
	}

	if promotionAt == nil {
		now := time.Now().UTC()
		promotionAt = &now
	}

	var _stats ConfigStat
	tx = tx.Debug().Table(ploutos.CashMethodPromotion{}.TableName()).
		Select("max(payout_rate) payout_rate_max, max(daily_max_payout) daily_max_payout_max, max(weekly_max_payout) weekly_max_payout_max, min(min_payout) min_payout_min").
		Where("cash_method_id", cashMethodId).Where("vip_id", vipId).
		Where("start_at < ? and end_at > ?", promotionAt, promotionAt).
		Where("status = ?", 1).
		Find(&_stats)
	err := tx.Error
	if err != nil {
		return ConfigStat{}, err
	}
	return _stats, nil
}

// FinalPossiblePayout
// cashAmount == nil => calculate ceiling for the payout
func FinalPossiblePayout(ctx context.Context, claimedPast7Days int64, claimedPast1Day int64, cashMethodPromotion ploutos.CashMethodPromotion, cashAmount *int64) (payout int64, err error) {
	ctx = rfcontext.AppendCallDesc(ctx, "FinalPossiblePayout")
	if claimedPast7Days >= cashMethodPromotion.WeeklyMaxPayout {
		log.Println(rfcontext.FmtJSON(rfcontext.AppendDescription(ctx, "weekly payout reached")))
		return
	}
	if claimedPast1Day >= cashMethodPromotion.DailyMaxPayout {
		log.Println(rfcontext.FmtJSON(rfcontext.AppendDescription(ctx, "daily payout reached")))
		return
	}

	dailyClaimableCeiling := float64(cashMethodPromotion.DailyMaxPayout - claimedPast1Day)
	weeklyClaimableCeiling := float64(cashMethodPromotion.WeeklyMaxPayout - claimedPast7Days)

	if cashAmount == nil {
		return int64(min(dailyClaimableCeiling, weeklyClaimableCeiling)), nil
	} else {
		ratedPayout := cashMethodPromotion.PayoutRate * float64(*cashAmount)
		return int64(min(ratedPayout, dailyClaimableCeiling, weeklyClaimableCeiling)), nil
	}
}
