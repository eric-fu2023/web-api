package model

import (
	"math"
	"time"

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
	err = DB.Transaction(func(tx *gorm.DB) error {
		// Use Count to get the number of records matching the conditions
		return tx.Table("taya_bet_report").
			Where("business_id = ?", businessId).
			Count(&count).Error
	})

	if count > 0 {
		isFoundBetReport = true
	}

	return
}
