package model

import (
	"time"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

type Teamup struct {
	ploutos.Teamup
}

type TeamupCustomRes []struct {
	UserId              string        `json:"user_id"`
	OrderId             string        `json:"order_id"`
	TotalTeamupDeposit  float64       `json:"total_teamup_deposit"`
	TotalTeamupTarget   float64       `json:"total_teamup_target"`
	TeamupProgress      float64       `json:"teamup_progress"`
	TeamupEndTime       time.Time     `json:"teamup_end_time"`
	TeamupCompletedTime time.Time     `json:"teamup_completed_time"`
	InfoJson            []byte        `json:"info_json,omitempty"`
	GameType            int64         `json:"game_type"`
	Bets                []ploutos.Bet `json:"bets"`
	BetList             []OutgoingBet `json:"bet_list"`
}

type OutgoingBet struct {
	MarketName   string `json:"market_name"`
	OptionName   string `json:"option_name"`
	MatchName    string `json:"match_name"`
	MatchTime    string `json:"match_time"`
	HomeName     string `json:"home_name"`
	AwayName     string `json:"away_name"`
	HomeIcon     string `json:"home_icon"`
	AwayIcon     string `json:"away_icon"`
	SettleResult *int64 `json:"settle_result"`
	IsInplay     *bool  `json:"is_inplay"`
	BetScore     string `json:"bet_score"`
	ExtraInfo    string `json:"extra_info"`
}

func SaveTeamup(teamup ploutos.Teamup) (err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Save(&teamup).Error
		return
	})

	return
}

func GetTeamUpBetReport(orderId string) (betReport ploutos.BetReport, err error) {

	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Where("business_id = ?", orderId).First(&betReport).Error
		return
	})

	return
}

func GetTeamUp(orderId string) (teamup ploutos.Teamup, err error) {

	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Model(ploutos.Teamup{}).Where("order_id = ?", orderId).First(&teamup).Error
		return
	})

	return
}

func GetAllTeamUps(userId int64, status []int, page, limit int) (res TeamupCustomRes, err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		tx = tx.Table("teamups").Select("teamups.user_id as user_id, teamups.total_teamup_deposit, teamups.total_teamup_target, teamups.teamup_end_time, teamups.teamup_completed_time, bet_report.business_id as order_id, bet_report.info as info_json, bet_report.game_type").
			Joins("left join bet_report on teamups.order_id = bet_report.business_id")

		tx = tx.Where("teamups.user_id = ?", userId)

		if len(status) > 0 {
			tx = tx.Where("teamups.status in ?", status)
		}

		err = tx.Scopes(Paginate(page, limit)).Scan(&res).Error
		return
	})

	return
}
