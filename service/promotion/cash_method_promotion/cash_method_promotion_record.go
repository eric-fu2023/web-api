package cash_method_promotion

import (
	"context"
	"time"

	"web-api/model"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

func PromotionRecordByCashOrderId(cashOrderId string, tx *gorm.DB) (cashMethodPromotionRecord ploutos.CashMethodPromotionRecord, err error) {
	if tx == nil {
		tx = model.DB
	}
	err = tx.Where("cash_order_id", cashOrderId).Find(&cashMethodPromotionRecord).Error
	return
}

type _ = ploutos.CashMethodPromotionRecord
type CashMethodPromotionRecordStats struct {
	ploutos.BASE
	CashMethodPromotionId int64 `json:"cash_method_promotion_id" form:"cash_method_promotion_id" gorm:"column:cash_method_promotion_id;comment:;size:64;"`
	CashMethodId          int64 `json:"cash_method_id" form:"cash_method_id" gorm:"column:cash_method_id;comment:;size:64;"`
	VipId                 int64 `json:"vip_id" form:"vip_id" gorm:"column:vip_id;comment:;size:64;"`

	UserId     int64   `json:"user_id" form:"user_id" gorm:"column:user_id;comment:;size:64;"`
	PayoutRate float64 `json:"payout_rate" form:"payout_rate" gorm:"column:payout_rate;comment:;"`
	Amount     int64   `json:"amount" form:"amount" gorm:"column:amount;comment:;size:64;"`
}

func (CashMethodPromotionRecordStats) TableName() string {
	return ploutos.TableNameCashMethodPromotionRecord
}

// totalClaimedByUserInPeriod
// FIXME this query returns a single aggregate tuple, not CashMethodPromotionRecord(s).
func totalClaimedByUserInPeriod(cashMethodId *int64, userId int64, startAt, endAt time.Time, tx *gorm.DB) (cashMethodPromotionRecords []CashMethodPromotionRecordStats, err error) {
	if tx == nil {
		tx = model.DB
	}

	tx = tx.Debug().
		Select("SUM(amount) as amount, cash_method_id, user_id").
		Group("cash_method_id").
		Group("user_id")

	if !startAt.IsZero() {
		tx = tx.Where("created_at >= ?", startAt)
	}
	if !endAt.IsZero() {
		tx = tx.Where("created_at < ?", endAt)
	}

	if cashMethodId != nil {
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

func GetAccumulatedClaimedCashMethodPromotionPast7And1Days(c context.Context, cashMethodId *int64, userId int64) (weeklyAmountRecords []CashMethodPromotionRecordStats, dailyAmountRecords []CashMethodPromotionRecordStats, err error) {
	now := time.Now()
	weeklyAmountRecords, err = totalClaimedByUserInPeriod(cashMethodId, userId, now.AddDate(0, 0, -7), now, nil)
	if err != nil {
		util.GetLoggerEntry(c).Error("GetAccumulatedClaimedCashMethodPromotionPast7And1Days totalClaimedByUserInPeriod", cashMethodId, userId)
		return
	}
	dailyAmountRecords, err = totalClaimedByUserInPeriod(cashMethodId, userId, now.AddDate(0, 0, -1), now, nil)
	if err != nil {
		util.GetLoggerEntry(c).Error("GetAccumulatedClaimedCashMethodPromotionPast7And1Days totalClaimedByUserInPeriod", cashMethodId, userId)
		return
	}
	util.GetLoggerEntry(c).Info("GetAccumulatedClaimedCashMethodPromotionPast7And1Days weeklyAmountRecords", weeklyAmountRecords, cashMethodId, userId, now.AddDate(0, 0, -7), now) // wl: for staging debug
	util.GetLoggerEntry(c).Info("GetAccumulatedClaimedCashMethodPromotionPast7And1Days dailyAmountRecords", dailyAmountRecords, cashMethodId, userId, now.AddDate(0, 0, -1), now)   // wl: for staging debug
	return
}

// GetAccumulatedClaimedCashMethodPromotionPast7And1DaysM
// aggregated by cash method
func GetAccumulatedClaimedCashMethodPromotionPast7And1DaysM(c context.Context, userId int64) (map[ /*cash method id*/ int64]CashMethodPromotionRecordStats, map[ /* cash method id */ int64]CashMethodPromotionRecordStats, error) {
	now := time.Now()
	claimed7DaysByCashMethod, err := totalClaimedByUserInPeriod(nil, userId, now.AddDate(0, 0, -7), now, nil)
	if err != nil {
		return make(map[ /*cash method id*/ int64]CashMethodPromotionRecordStats, 0), make(map[ /*cash method id*/ int64]CashMethodPromotionRecordStats, 0), err
	}
	claimed1DayByCashMethod, err := totalClaimedByUserInPeriod(nil, userId, now.AddDate(0, 0, -1), now, nil)
	if err != nil {
		return make(map[ /*cash method id*/ int64]CashMethodPromotionRecordStats, 0), make(map[ /*cash method id*/ int64]CashMethodPromotionRecordStats, 0), err
	}

	claimed7DaysByCashMethodMap := make(map[ /*cash method id*/ int64]CashMethodPromotionRecordStats, 0)

	for _, v := range claimed7DaysByCashMethod {
		claimed7DaysByCashMethodMap[v.CashMethodId] = v
	}

	claimed1DayByCashMethodMap := make(map[ /*cash method id*/ int64]CashMethodPromotionRecordStats, 0)
	for _, v := range claimed1DayByCashMethod {
		claimed1DayByCashMethodMap[v.CashMethodId] = v
	}

	return claimed7DaysByCashMethodMap, claimed1DayByCashMethodMap, nil
}
