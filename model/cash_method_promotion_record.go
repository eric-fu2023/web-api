package model

import (
	"context"
	"time"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

func FindCashMethodPromotionRecordByCashOrderId(cashOrderId string, tx *gorm.DB) (cashMethodPromotionRecord models.CashMethodPromotionRecord, err error) {
	if tx == nil {
		tx = DB
	}
	err = tx.Where("cash_order_id", cashOrderId).Find(&cashMethodPromotionRecord).Error
	if err != nil {
		return
	}
	return
}

func AggreCashMethodPromotionRecordAmountByCashMethodIdAndUserId(cashMethodId, userId int64, startAt, endAt time.Time, tx *gorm.DB) (cashMethodPromotionRecords []models.CashMethodPromotionRecord, err error) {
	if tx == nil {
		tx = DB
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

func GetWeeklyAndDailyCashMethodPromotionRecord(c context.Context, cashMethodId, userId int64) (weeklyAmountRecords, dailyAmountRecords []models.CashMethodPromotionRecord, err error) {
	now := time.Now()
	weeklyAmountRecords, err = AggreCashMethodPromotionRecordAmountByCashMethodIdAndUserId(cashMethodId, userId, now.AddDate(0, 0, -7), now, nil)
	if err != nil {
		util.GetLoggerEntry(c).Error("GetWeeklyAndDailyCashMethodPromotionRecord AggreCashMethodPromotionRecordAmountByCashMethodIdAndUserId", cashMethodId, userId)
		return
	}
	dailyAmountRecords, err = AggreCashMethodPromotionRecordAmountByCashMethodIdAndUserId(cashMethodId, userId, now.AddDate(0, 0, -1), now, nil)
	if err != nil {
		util.GetLoggerEntry(c).Error("GetWeeklyAndDailyCashMethodPromotionRecord AggreCashMethodPromotionRecordAmountByCashMethodIdAndUserId", cashMethodId, userId)
		return
	}
	util.GetLoggerEntry(c).Info("GetWeeklyAndDailyCashMethodPromotionRecord weeklyAmountRecords", weeklyAmountRecords, cashMethodId, userId, now.AddDate(0, 0, -7), now) // wl: for staging debug
	util.GetLoggerEntry(c).Info("GetWeeklyAndDailyCashMethodPromotionRecord dailyAmountRecords", dailyAmountRecords, cashMethodId, userId, now.AddDate(0, 0, -1), now)   // wl: for staging debug
	return
}
