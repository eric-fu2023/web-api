package model

import (
	"fmt"
	"sort"
	"time"

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
	TeamupEndTime       int64   `json:"teamup_end_time"`
	TeamupLocalEndTime  int64   `json:"teamup_local_end_time"` // Use carefully
	TeamupCompletedTime int64   `json:"teamup_completed_time"`
	InfoJson            []byte  `json:"info_json,omitempty"`
	GameType            int64   `json:"game_type"`
	IsParlay            bool    `json:"is_parlay"`
	BetType             string  `json:"bet_type"`
	TotalFakeProgress   int64   `json:"total_fake_progress"`
	LeagueIcon          string  `json:"league_icon"`
	LeagueName          string  `json:"league_name"`
	HomeIcon            string  `json:"home_icon"`
	HomeName            string  `json:"home_name"`
	AwayIcon            string  `json:"away_icon"`
	AwayName            string  `json:"away_name"`
	Status              int64   `json:"status"`

	MarketName string `json:"market_name"`
	OptionName string `json:"option_name"`
	MatchTitle string `json:"match_title"`
	MatchId    string `json:"match_id"`
	MatchTime  int64  `json:"match_time"`

	BetReportGameType int `json:"bet_report_game_type"`
}

type OutgoingTeamupCustomRes []struct {
	TeamupId            string      `json:"teamup_id"`
	UserId              string      `json:"user_id"`
	OrderId             string      `json:"order_id"`
	TotalTeamupDeposit  float64     `json:"total_teamup_deposit"`
	TotalTeamupTarget   float64     `json:"total_teamup_target"`
	TeamupEndTime       int64       `json:"teamup_end_time"`
	TeamupLocalEndTime  int64       `json:"teamup_local_end_time"` // Use carefully
	TeamupCompletedTime int64       `json:"teamup_completed_time"`
	InfoJson            []byte      `json:"info_json,omitempty"`
	GameType            int64       `json:"game_type"`
	IsParlay            bool        `json:"is_parlay"`
	BetType             string      `json:"bet_type"`
	Bet                 OutgoingBet `json:"bet"`
	TotalFakeProgress   int64       `json:"total_fake_progress"`
	LeagueIcon          string      `json:"league_icon"`
	LeagueName          string      `json:"league_name"`
	HomeIcon            string      `json:"home_icon"`
	HomeName            string      `json:"home_name"`
	AwayIcon            string      `json:"away_icon"`
	AwayName            string      `json:"away_name"`
	Status              int64       `json:"status"`
	HasJoined           bool        `json:"has_joined"`

	MarketName string `json:"market_name"`
	OptionName string `json:"option_name"`
	MatchTitle string `json:"match_title"`
	MatchId    string `json:"match_id"`
	MatchTime  int64  `json:"match_time"`

	BetReportGameType int `json:"bet_report_game_type"`
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

type TeamupSuccess []struct {
	Nickname string `json:"nickname"`
	Time     int64  `json:"time"`
	Amount   int64  `json:"amount"`
	Avatar   string `json:"avatar"`
	IsReal   bool   `json:"is_real"`
}

func SaveTeamup(teamup ploutos.Teamup) (t ploutos.Teamup, err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Save(&teamup).Error
		t = teamup
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

// teamups.match_id as match_id, teamups.match_time as match_time, teamups.total_fake_progress as total_fake_progress,
// tx = tx.Table("teamups").Select("teamups.league_name, teamups.option_name, teamups.bet_report_game_type, teamups.market_name, teamups.option_name, teamups.is_parlay, teamups.match_title, teamups.match_id, teamups.match_time, teamups.status, teamups.league_icon, teamups.home_icon, teamups.away_icon, teamups.total_fake_progress, teamups.id as teamup_id, teamups.user_id as user_id, teamups.total_teamup_deposit, teamups.total_teamup_target, teamups.teamup_end_time, teamups.teamup_completed_time, bet_report.business_id as order_id, bet_report.info as info_json, bet_report.game_type, bet_report.is_parlay, bet_report.bet_type").
func GetAllTeamUps(userId int64, status []int, page, limit int, start, end int64) (res TeamupCustomRes, err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		tx = tx.Table("teamups").Select("teamups.bet_report_game_type as bet_report_game_type, teamups.match_id as match_id, teamups.match_time as match_time, teamups.total_fake_progress as total_fake_progress, teamups.match_title as match_title, teamups.league_icon as league_icon, teamups.home_icon as home_icon, teamups.away_icon as away_icon, teamups.status as status, teamups.league_name as league_name, teamups.option_name as option_name, teamups.market_name as market_name, teamups.is_parlay as is_parlay, teamups.id as teamup_id, teamups.user_id as user_id, teamups.total_teamup_deposit, teamups.total_teamup_target, teamups.teamup_end_time, teamups.teamup_completed_time, teamups.order_id as order_id, teamups.is_parlay as is_parlay, teamups.match_title as bet_type")

		tx = tx.Where("teamups.user_id = ?", userId)

		if len(status) > 0 {
			tx = tx.Where("teamups.status in ?", status)
		}

		if start != 0 && end != 0 {
			tx = tx.Where(`teamup_end_time >= ?`, start).Where(`teamup_end_time <= ?`, end)
		}

		switch {
		case len(status) == 1:
			tx = tx.Order(`teamup_end_time ASC`)
		case len(status) == 2:
			tx = tx.Order(`teamup_end_time DESC`)
		case len(status) == 3:
			tx = tx.Order(`teamups.status ASC`).Order(`teamup_end_time ASC`)
		}

		err = tx.Scopes(Paginate(page, limit)).Scan(&res).Error

		// 成功和失败按照时间排序
		endedStartIndex := 0
		endedEndIndex := len(res) - 1

		if len(status) == 3 {
			for i, t := range res {
				if t.Status != 0 {
					endedStartIndex = i
					break
				}
			}

			if endedStartIndex >= 0 && endedEndIndex < len(res) && endedStartIndex <= endedEndIndex {
				sort.Slice(res[endedStartIndex:endedEndIndex+1], func(i, j int) bool {
					return res[endedStartIndex+i].TeamupEndTime > res[endedStartIndex+j].TeamupEndTime
				})
			}
		}

		return
	})

	if err != nil {
		return
	}

	return
}

func GetCustomTeamUpByTeamUpId(teamupId int64) (res TeamupCustomRes, err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		tx = tx.Table("teamups").Select("teamups.league_name, teamups.option_name, teamups.bet_report_game_type, teamups.market_name, teamups.option_name, teamups.is_parlay, teamups.match_title, teamups.match_id, teamups.match_time, teamups.status, teamups.league_icon, teamups.home_icon, teamups.away_icon, teamups.total_fake_progress, teamups.id as teamup_id, teamups.user_id as user_id, teamups.total_teamup_deposit, teamups.total_teamup_target, teamups.teamup_end_time, teamups.teamup_completed_time, teamups.order_id as order_id, teamups.is_parlay as is_parlay, teamups.match_title as bet_type")

		tx = tx.Where("teamups.id = ?", teamupId)

		err = tx.Scan(&res).Error
		return
	})
	if err != nil {
		return
	}

	return
}

func GetTeamUpByTeamUpId(teamupId int64) (res ploutos.Teamup, err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Where("id = ?", teamupId).First(&res).Error
		return
	})

	return
}

func GetTeamupProgressToUpdate(userId, amount, slashProgress int64) (err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		teamupEntry, err := FindOngoingTeamupEntriesByUserId(userId)
		if err != nil {
			err = fmt.Errorf("fail to get teamup err - %v", err)
			return
		}

		err = UpdateFirstTeamupEntryProgress(tx, teamupEntry.ID, amount, slashProgress)

		if err != nil {
			err = fmt.Errorf("fail to update teamup entry err - %v", err)
			return
		}

		err = UpdateTeamupProgress(tx, teamupEntry.TeamupId, amount, slashProgress)

		if err != nil {
			err = fmt.Errorf("fail to update teamup err - %v", err)
			return
		}
		return
	})

	return
}

func UpdateTeamupProgress(tx *gorm.DB, teamupId, amount, slashAmount int64) error {
	return tx.Transaction(func(tx2 *gorm.DB) error {

		updates := map[string]interface{}{
			"total_accumulated_deposit": gorm.Expr("total_accumulated_deposit + ?", amount),
			"total_teamup_deposit":      gorm.Expr("total_teamup_deposit + ?", slashAmount),
		}

		if err := tx2.Table("teamups").
			Where("id = ?", teamupId).
			Limit(1).
			Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})
}

func GetRecentCompletedSuccessTeamup(numMinutes int64) (res TeamupSuccess, err error) {

	recentMinutes := time.Duration(numMinutes) * time.Minute

	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		tx = tx.Table("teamups").Select("users.nickname, teamups.teamup_completed_time as time, users.avatar, teamups.total_teamup_target as amount").
			Joins("left join users on teamups.user_id = users.id")

		tx = tx.Where("teamups.status = 1").Where("teamups.teamup_completed_time >= ?", time.Now().UTC().Unix()-int64(recentMinutes.Seconds()))

		err = tx.Scan(&res).Error
		return
	})
	if err != nil {
		return
	}

	return
}
