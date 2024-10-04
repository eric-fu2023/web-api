package cash_method_promotion

import (
	"context"
	"time"

	"web-api/model"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

func PromotionRecordByCashOrderId(cashOrderId string, tx *gorm.DB) (cashMethodPromotionRecord models.CashMethodPromotionRecord, err error) {
	if tx == nil {
		tx = model.DB
	}
	err = tx.Where("cash_order_id", cashOrderId).Find(&cashMethodPromotionRecord).Error
	if err != nil {
		return
	}
	return
}

// TotalClaimedByUserInPeriod
// FIXME this query returns a single aggregate tuple, not CashMethodPromotionRecord(s).
func TotalClaimedByUserInPeriod(cashMethodId, userId int64, startAt, endAt time.Time, tx *gorm.DB) (cashMethodPromotionRecords []models.CashMethodPromotionRecord, err error) {
	if tx == nil {
		tx = model.DB
	}

	tx = tx.
		Select("SUM(amount) as amount, cash_method_id, user_id").
		Group("cash_method_id").
		Group("user_id")

	if !startAt.IsZero() {
		tx = tx.Where("created_at >= ?", startAt)
	}
	if !endAt.IsZero() {
		tx = tx.Where("created_at < ?", endAt)
	}

	if cashMethodId != 0 {
		tx = tx.Where("cash_method_id", cashMethodId)
	}
	if userId != 0 {
		tx = tx.Where("user_id", userId)
	}

	err = tx.Find(&cashMethodPromotionRecords).Error
	if err != nil {
		return
	}

	return
}

func GetAccumulatedClaimedCashMethodPromotionPast7And1Days(c context.Context, cashMethodId, userId int64) (weeklyAmountRecords []models.CashMethodPromotionRecord, dailyAmountRecords []models.CashMethodPromotionRecord, err error) {
	now := time.Now()
	weeklyAmountRecords, err = TotalClaimedByUserInPeriod(cashMethodId, userId, now.AddDate(0, 0, -7), now, nil)
	if err != nil {
		util.GetLoggerEntry(c).Error("GetAccumulatedClaimedCashMethodPromotionPast7And1Days TotalClaimedByUserInPeriod", cashMethodId, userId)
		return
	}
	dailyAmountRecords, err = TotalClaimedByUserInPeriod(cashMethodId, userId, now.AddDate(0, 0, -1), now, nil)
	if err != nil {
		util.GetLoggerEntry(c).Error("GetAccumulatedClaimedCashMethodPromotionPast7And1Days TotalClaimedByUserInPeriod", cashMethodId, userId)
		return
	}
	util.GetLoggerEntry(c).Info("GetAccumulatedClaimedCashMethodPromotionPast7And1Days weeklyAmountRecords", weeklyAmountRecords, cashMethodId, userId, now.AddDate(0, 0, -7), now) // wl: for staging debug
	util.GetLoggerEntry(c).Info("GetAccumulatedClaimedCashMethodPromotionPast7And1Days dailyAmountRecords", dailyAmountRecords, cashMethodId, userId, now.AddDate(0, 0, -1), now)   // wl: for staging debug
	return
}
