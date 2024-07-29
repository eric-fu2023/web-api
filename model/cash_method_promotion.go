package model

import (
	"context"
	"math"
	"time"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

func FindActiveCashMethodPromotionByCashMethodIdAndVipId(cashMethodId, vipId int64, promotionAt *time.Time, tx *gorm.DB) (cashMethodPromotion models.CashMethodPromotion, err error) {
	if promotionAt == nil {
		now := time.Now().UTC()
		promotionAt = &now
	}
	if tx == nil {
		tx = DB
	}
	err = tx.
		Where("cash_method_id", cashMethodId).Where("vip_id", vipId).
		Where("start_at < ? and end_at > ?", promotionAt, promotionAt).Where("status = ?", 1).
		Find(&cashMethodPromotion).Error
	if err != nil {
		return
	}
	return
}

func GetMaxCashMethodPromotionAmount(c context.Context, weeklyAmount, dailyAmount int64, cashMethodPromotion models.CashMethodPromotion, userId, cashAmount int64, noCashAmount bool) (amount int64, err error) {
	if weeklyAmount >= cashMethodPromotion.WeeklyMaxPayout {
		util.GetLoggerEntry(c).Info("GetMaxCashMethodPromotionAmount weeklyAmount >= cashMethodPromotion.WeeklyMaxPayout", weeklyAmount, cashMethodPromotion.WeeklyMaxPayout)
		return
	}
	if dailyAmount >= cashMethodPromotion.DailyMaxPayout {
		util.GetLoggerEntry(c).Info("GetMaxCashMethodPromotionAmount dailyAmount >= cashMethodPromotion.DailyMaxPayout", dailyAmount, cashMethodPromotion.DailyMaxPayout)
		return
	}

	oriAmount := cashMethodPromotion.PayoutRate * float64(cashAmount)
	maxDailyPayoutRemaining := float64(cashMethodPromotion.DailyMaxPayout - dailyAmount)
	maxWeeklyPayoutRemaining := float64(cashMethodPromotion.WeeklyMaxPayout - weeklyAmount)

	if !noCashAmount {
		amount = int64(math.Min(oriAmount, maxDailyPayoutRemaining))
	} else {
		amount = int64(maxDailyPayoutRemaining)
	}
	amount = int64(math.Min(float64(amount), maxWeeklyPayoutRemaining))
	util.GetLoggerEntry(c).Info("GetMaxCashMethodPromotionAmount get min amount", oriAmount, maxDailyPayoutRemaining, maxWeeklyPayoutRemaining, amount) // wl: for staging debug
	return
}
