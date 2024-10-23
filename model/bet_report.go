package model

import (
	"context"
	"fmt"
	"math"
	"time"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

type ReportDetail struct {
	TotalNegativeProfit int64 `json:"total_negative_profit"`
}

func GetNegativeProfitRebate(startDate, endDate time.Time, userId, percentage int64) (res int64, err error) {

	// If totalNegativeProfit > 0, means user bet result overall lose = will give them rebate

	// const REVIVE_REBATE_PERCENTAGE = 12

	totalNegativeProfit, err := GetNegativeProfit(startDate, endDate, userId)

	res = int64(math.Floor((float64(totalNegativeProfit) / float64(100)) * float64(percentage)))

	return
}

func GetNegativeProfit(startDate, endDate time.Time, userId int64) (res int64, err error) {
	// SUM ALL ONLY SPORTS BETTING (FB, TAYA, IMSB)
	// If totalNegativeProfit > 0, means user bet result overall lose

	var reportDetail ReportDetail
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Table("bet_report").
			Select("SUM(profit_loss) as total_negative_profit").
			Where("user_id = ?", userId).
			Where("status = 5").
			Where("game_type in (?)", []int{ploutos.GAME_FB, ploutos.GAME_TAYA, ploutos.GAME_IMSB}).
			Scopes(Range(startDate, endDate, "reward_time")).
			Scan(&reportDetail).Error

		if err != nil {
			return
		}
		if reportDetail.TotalNegativeProfit >= 0 {
			res = reportDetail.TotalNegativeProfit
		}
		return
	})

	return
}

func GetBetReportByBusinessId(businessId string) (isFoundBetReport bool, err error) {
	var count int64
	err = DB.Table("bet_report").
		Where("business_id = ?", businessId).
		Count(&count).Error

	if count > 0 {
		isFoundBetReport = true
	}

	return
}

func GetBetReport(businessId string) (betReport ploutos.BetReport, err error) {
	err = DB.Table("bet_report").
		Where("business_id = ?", businessId).
		First(&betReport).Error
	return
}

type OrderSummary struct {
	Count  int64 `gorm:"column:count"`
	Amount int64 `gorm:"column:amount"`
	Win    int64 `gorm:"column:win"`
}

func BetReportsStats(ctx context.Context, userId int64, fromBetTime, toBetTime time.Time, gameVendorIds []int64, statuses []ploutos.TayaBetReportStatus, isParlay bool) (OrderSummary, error) {
	var orderSummary OrderSummary
	ctx = rfcontext.AppendCallDesc(ctx, "CountBetReports")
	db := DB
	if db == nil {
		return OrderSummary{}, fmt.Errorf("db is nil")
	}
	err := db.Model(BetReport{}).Debug().Select(`COUNT(1) as count, SUM(bet) as amount, SUM(win-bet) as win`).Scopes(ByOrderListConditions(userId, gameVendorIds, statuses, &isParlay, fromBetTime, toBetTime)).Find(&orderSummary).Error

	if err != nil {
		return OrderSummary{}, err
	}
	return orderSummary, err
}

func BetReports(ctx context.Context, userId int64, fromBetTime, toBetTime time.Time, gameVendorIds []int64, statusesToInclude []ploutos.TayaBetReportStatus, isParlay bool, pageNo int, pageSize int) ([]ploutos.BetReport, error) {
	ctx = rfcontext.AppendCallDesc(ctx, "BetReports")
	db := DB
	if db == nil {
		return []ploutos.BetReport{}, fmt.Errorf("db is nil")
	}

	var betReports []ploutos.BetReport
	err := DB.Preload(`Voucher`).Preload(`ImVoucher`).Preload(`GameVendor`).
		Model(ploutos.BetReport{}).Debug().Scopes(ByOrderListConditions(userId, gameVendorIds, statusesToInclude, &isParlay, fromBetTime, toBetTime), ByBetTimeSort, Paginate(pageNo, pageSize)).
		Find(&betReports).Error

	if err != nil {
		return []ploutos.BetReport{}, err
	}
	return betReports, nil
}

func IsSettledFlagToPloutosIncludeStatuses(s *bool, forAggregation bool) []ploutos.TayaBetReportStatus {
	var statuses []ploutos.TayaBetReportStatus
	if s == nil { // "default"
		statuses = []ploutos.TayaBetReportStatus{ploutos.TayaBetReportStatusCreated, ploutos.TayaBetReportStatusConfirming,
			ploutos.TayaBetReportStatusRejected,
			ploutos.TayaBetReportStatusCancelled,
			ploutos.TayaBetReportStatusConfirmed,
			ploutos.TayaBetReportStatusSettled, 6}
	} else if /*service.IsSettled != nil &&*/ *s {
		if forAggregation { // only include effective bet amounts
			statuses = []ploutos.TayaBetReportStatus{ploutos.TayaBetReportStatusSettled, 6}
		} else {
			statuses = []ploutos.TayaBetReportStatus{ploutos.TayaBetReportStatusRejected,
				ploutos.TayaBetReportStatusCancelled, ploutos.TayaBetReportStatusSettled, 6}
		}
	} else /*service.IsSettled != nil && !*service.IsSettled */ {
		statuses = []ploutos.TayaBetReportStatus{ploutos.TayaBetReportStatusCreated,
			ploutos.TayaBetReportStatusConfirming,
			ploutos.TayaBetReportStatusConfirmed}
	}

	return statuses
}
