package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

type Teamup struct {
	ploutos.Teamup
}

type TeamupCustomRes []struct {
	TeamupId            string  `json:"teamup_id"`
	UserId              string  `json:"user_id"`
	OrderId             string  `json:"order_id"`
	TotalTeamupDeposit  float64 `json:"total_teamup_deposit"`
	TotalTeamupTarget   float64 `json:"total_teamup_target"`
	TeamupProgress      float64 `json:"teamup_progress"`
	TeamupEndTime       int64   `json:"teamup_end_time"`
	TeamupCompletedTime int64   `json:"teamup_completed_time"`
	InfoJson            []byte  `json:"info_json,omitempty"`
	GameType            int64   `json:"game_type"`
	IsParlay            bool    `json:"is_parlay"`
	BetType             string  `json:"bet_type"`
	TotalFakeProgress   int64   `json:"total_fake_progress"`
	LeagueIcon          string  `json:"league_icon"`
	HomeIcon            string  `json:"home_icon"`
	AwayIcon            string  `json:"away_icon"`
	Status              int64   `json:"status"`
}

type OutgoingTeamupCustomRes []struct {
	TeamupId            string      `json:"teamup_id"`
	UserId              string      `json:"user_id"`
	OrderId             string      `json:"order_id"`
	TotalTeamupDeposit  float64     `json:"total_teamup_deposit"`
	TotalTeamupTarget   float64     `json:"total_teamup_target"`
	TeamupProgress      float64     `json:"teamup_progress"`
	TeamupEndTime       int64       `json:"teamup_end_time"`
	TeamupCompletedTime int64       `json:"teamup_completed_time"`
	InfoJson            []byte      `json:"info_json,omitempty"`
	GameType            int64       `json:"game_type"`
	IsParlay            bool        `json:"is_parlay"`
	BetType             string      `json:"bet_type"`
	Bet                 OutgoingBet `json:"bet"`
	TotalFakeProgress   int64       `json:"total_fake_progress"`
	LeagueIcon          string      `json:"league_icon"`
	HomeIcon            string      `json:"home_icon"`
	AwayIcon            string      `json:"away_icon"`
	Status              int64       `json:"status"`
}

type OutgoingBet struct {
	MatchId      string `json:"match_id"`
	MarketName   string `json:"market_name"`
	LeagueName   string `json:"league_name"`
	LeagueIcon   string `json:"league_icon"`
	OptionName   string `json:"option_name"`
	MatchName    string `json:"match_name"`
	MatchTime    int64  `json:"match_time"`
	HomeName     string `json:"home_name"`
	AwayName     string `json:"away_name"`
	HomeIcon     string `json:"home_icon"`
	AwayIcon     string `json:"away_icon"`
	SettleResult int64  `json:"settle_result"`
	IsInplay     bool   `json:"is_inplay"`
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

func GetAllTeamUps(userId int64, status []int, page, limit int, start, end int64) (res TeamupCustomRes, err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		tx = tx.Table("teamups").Select("teamups.league_icon, teamups.home_icon, teamups.away_icon, teamups.status, teamups.id as teamup_id, teamups.user_id as user_id, teamups.total_teamup_deposit, teamups.total_teamup_target, teamups.teamup_end_time, teamups.teamup_completed_time, bet_report.business_id as order_id, bet_report.info as info_json, bet_report.game_type, bet_report.is_parlay, bet_report.bet_type").
			Joins("left join bet_report on teamups.order_id = bet_report.business_id")

		tx = tx.Where("teamups.user_id = ?", userId)

		if len(status) > 0 {
			tx = tx.Where("teamups.status in ?", status)
		}

		if start != 0 && end != 0 {
			tx = tx.Where(`teamup_end_time >= ?`, start).Where(`teamup_end_time <= ?`, end)
		}

		err = tx.Scopes(Paginate(page, limit)).Scan(&res).Error
		return
	})

	return
}

func GetCustomTeamUpByTeamUpId(teamupId int64) (res TeamupCustomRes, err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		tx = tx.Table("teamups").Select("teamups.league_icon, teamups.home_icon, teamups.away_icon, teamups.total_fake_progress, teamups.id as teamup_id, teamups.user_id as user_id, teamups.total_teamup_deposit, teamups.total_teamup_target, teamups.teamup_end_time, teamups.teamup_completed_time, bet_report.business_id as order_id, bet_report.info as info_json, bet_report.game_type, bet_report.is_parlay, bet_report.bet_type").
			Joins("left join bet_report on teamups.order_id = bet_report.business_id")

		tx = tx.Where("teamups.id = ?", teamupId)

		err = tx.Scan(&res).Error
		return
	})

	return
}

func GetTeamUpByTeamUpId(teamupId int64) (res ploutos.Teamup, err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Where("id = ?", teamupId).First(&res).Error
		return
	})

	return
}
