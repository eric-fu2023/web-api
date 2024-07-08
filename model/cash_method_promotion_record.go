package model

import (
	"time"

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

func AggreCashMethodPromotionRecordAmountByCashMethodIdAndUserId(cashMethodId, userId int64, startAt, endAt time.Time, tx *gorm.DB) (amount int64, err error) {
	if tx == nil {
		tx = DB
	}
	var cashMethodPromotionRecord models.CashMethodPromotionRecord
	tx = tx.
		Select("SUM(amount) as amount").
		Where("cash_method_id", cashMethodId).
		Where("user_id", userId).
		Group("cash_method_id").
		Group("user_id")

	if !startAt.IsZero() {
		tx = tx.Where("created_at >= ?", startAt)

	}
	if !endAt.IsZero() {
		tx = tx.Where("created_at < ?", endAt)
	}

	err = tx.Find(&cashMethodPromotionRecord).Error
	if err != nil {
		return
	}

	amount = cashMethodPromotionRecord.Amount
	return
}
