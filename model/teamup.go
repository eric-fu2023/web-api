package model

import (
	"sort"
	"strconv"
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
	TeamupLocalEndTime  string  `json:"teamup_local_end_time"` // Use carefully
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
	Provider            string  `json:"provider"`
	TeamupType          int64   `json:"teamup_type"`

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
	TeamupLocalEndTime  string      `json:"teamup_local_end_time"` // Use carefully
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
	Provider            string      `json:"provider"`
	TeamupType          int64       `json:"teamup_type"`

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

func GetAllTeamUps(userId int64, status []int, page, limit int, start, end int64) (res TeamupCustomRes, err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		tx = tx.Table("teamups").Select("teamups.provider, teamups.bet_report_game_type as bet_report_game_type, teamups.match_id as match_id, teamups.match_time as match_time, teamups.total_fake_progress as total_fake_progress, teamups.match_title as match_title, teamups.league_icon as league_icon, teamups.home_icon as home_icon, teamups.away_icon as away_icon, teamups.status as status, teamups.league_name as league_name, teamups.option_name as option_name, teamups.market_name as market_name, teamups.is_parlay as is_parlay, teamups.id as teamup_id, teamups.user_id as user_id, teamups.total_teamup_deposit, teamups.total_teamup_target, teamups.teamup_end_time, teamups.teamup_completed_time, teamups.order_id as order_id, teamups.is_parlay as is_parlay, teamups.match_title as bet_type")

		tx = tx.Where("teamups.user_id = ?", userId)

		if len(status) > 0 {
			tx = tx.Where("teamups.status in ?", status)
		}

		if start != 0 && end != 0 {
			tx = tx.Where(`teamup_end_time >= ?`, start).Where(`teamup_end_time <= ?`, end)
		}

		switch {
		case len(status) == 1:
			tx = tx.Order(`teamup_end_time ASC`).Order(`created_at DESC`)
		case len(status) == 2:
			tx = tx.Order(`teamup_end_time DESC`).Order(`created_at DESC`)
		case len(status) == 3:
			// tx = tx.Order(`teamups.status ASC`).Order(`created_at DESC`).Order(`teamup_end_time ASC`)
			tx = tx.Order(`teamups.status ASC`).Order(`teamup_end_time ASC`).Order(`created_at DESC`)
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

			if endedStartIndex > 0 && endedEndIndex < len(res) && endedStartIndex <= endedEndIndex {
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
		tx = tx.Table("teamups").Select("teamups.provider, teamups.league_name, teamups.option_name, teamups.bet_report_game_type, teamups.market_name, teamups.option_name, teamups.is_parlay, teamups.match_title, teamups.match_id, teamups.match_time, teamups.status, teamups.league_icon, teamups.home_icon, teamups.away_icon, teamups.total_fake_progress, teamups.id as teamup_id, teamups.user_id as user_id, teamups.total_teamup_deposit, teamups.total_teamup_target, teamups.teamup_end_time, teamups.teamup_completed_time, teamups.order_id as order_id, teamups.is_parlay as is_parlay, teamups.match_title as bet_type")

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

		// tx = tx.Where("teamups.status = 1").Where("teamups.teamup_completed_time >= ?", time.Now().UTC().Unix()-int64(recentMinutes.Seconds())).
		// 	Where("teamups.bet_report_game_type in ?", ploutos.TeamUpSportGameTypes)

		tx = tx.Where("teamups.status = 1").Where("teamups.teamup_completed_time >= ?", time.Now().UTC().Unix()-int64(recentMinutes.Seconds()))

		err = tx.Scan(&res).Error
		return
	})
	if err != nil {
		return
	}

	return
}

func GetCurrentTermNum(brandId, gameType int) (maxTerm int64, err error) {

	gameTypes, _ := GetGameTypeSlice(brandId, gameType)

	err = DB.Table("teamups").
		Select("MAX(term)").
		Where("bet_report_game_type IN ?", gameTypes).
		Scan(&maxTerm).Error

	if err == nil && maxTerm == 0 {
		maxTerm = 1
	}

	return
}

func FindExceedTargetByTerm(brandId int, termId int64, gameType int) (teamups []ploutos.Teamup, err error) {

	gameTypes, _ := GetGameTypeSlice(brandId, gameType)

	err = DB.Table("teamups").
		Where("term = ?", termId).
		Where("bet_report_game_type IN ?", gameTypes).
		Find(&teamups).Error

	return
}

// 如果一个成功，同一届的候选池里其他砍单都自动失败，只有一个成功
func SuccessShortlisted(brandId int, teamup ploutos.Teamup, teamupEntriesCurrentProgress int64, finalSlashUserId int64) (isSuccess bool, err error) {

	maxPercentage := int64(10000)
	isSuccess = false

	if teamup.ShortlistStatus == ploutos.ShortlistStatusShortlistWin {
		return
	}

	hasWinnerAlready := false

	err = DB.Transaction(func(tx *gorm.DB) (err error) {

		gameTypes, _ := GetGameTypeSlice(brandId, teamup.BetReportGameType)

		// 如果该届/期已经有候选池里为成功的单子，不管砍单是否有成功都算成功
		// 可看下面注释
		var wonTeamups []ploutos.Teamup
		err = tx.Model(ploutos.Teamup{}).
			Where("term = ?", teamup.Term).
			Where("shortlist_status = ?", ploutos.ShortlistStatusShortlistWin).
			Where("bet_report_game_type IN ?", gameTypes).
			Find(&wonTeamups).Error

		if len(wonTeamups) > 0 {
			hasWinnerAlready = true
			return
		}

		teamup.ShortlistStatus = ploutos.ShortlistStatusShortlistWin

		// 如果砍单价值超过上限 砍单仍然还是进行中知道失败，但候选池里为成功所以不会用另一张成功的单子
		// 单子仍然进行中
		// 单子进度依旧，不会是100%
		teamupMaxSlashAmountString, _ := GetAppConfigWithCache("teamup", "max_slash_amount")
		teamupMaxSlashAmount, _ := strconv.Atoi(teamupMaxSlashAmountString)
		if teamup.TotalTeamUpTarget <= int64(teamupMaxSlashAmount) {
			teamup.Status = int(ploutos.TeamupStatusSuccess)
			teamup.TotalFakeProgress = maxPercentage
			teamup.TeamupCompletedTime = time.Now().UTC().Unix()
		}
		err = tx.Save(&teamup).Error
		if err != nil {
			return
		}

		slashEntry := ploutos.TeamupEntry{
			TeamupId: teamup.ID,
			UserId:   finalSlashUserId,
		}

		slashEntry.TeamupEndTime = teamup.TeamupEndTime
		slashEntry.TeamupCompletedTime = teamup.TeamupCompletedTime

		// 如果砍单价值超过上限 新用户贡献的进度为0
		if teamup.TotalTeamUpTarget <= int64(teamupMaxSlashAmount) {
			slashEntry.FakePercentageProgress = maxPercentage - teamupEntriesCurrentProgress
		}
		err = tx.Save(&slashEntry).Error

		return
	})
	if err != nil {
		return
	}

	if !hasWinnerAlready {
		isSuccess = true
	}
	return
}

func FlagStatusShortlistedWin(tx *gorm.DB, ids []int64) (err error) {
	return tx.Transaction(func(tx2 *gorm.DB) error {
		err = tx2.Model(ploutos.Teamup{}).
			Where("id IN ?", ids).
			Updates(map[string]interface{}{
				"shortlist_status":      ploutos.ShortlistStatusShortlistWin,
				"status":                int(ploutos.TeamupStatusSuccess),
				"total_fake_progress":   int64(10000),
				"teamup_completed_time": time.Now().UTC().Unix(),
			}).Error

		return err
	})
}

func GetGameTypeSlice(brandId, gameType int) (gameTypeIds []int, teamupType int) {

	teamupType = 0

	if brandId != int(BrandIdAha) && brandId != int(BrandIdBatace) {
		brandId = int(BrandIdBatace)
	}

	switch brandId {
	case int(BrandIdBatace):
		for _, t := range ploutos.TeamUpBASportGameTypes {
			if t == gameType {
				gameTypeIds = ploutos.TeamUpBASportGameTypes
				teamupType = int(ploutos.TeamupTypeSports)
			}
		}

		for _, t := range ploutos.TeamUpBAGameGameTypes {
			if t == gameType {
				gameTypeIds = ploutos.TeamUpBAGameGameTypes
				teamupType = int(ploutos.TeamupTypeGames)
			}
		}

		for _, t := range ploutos.TeamUpBASpinGameTypes {
			if t == gameType {
				gameTypeIds = ploutos.TeamUpBASpinGameTypes
				teamupType = int(ploutos.TeamupTypeSpin)
			}
		}

		return
	case int(BrandIdAha):
		for _, t := range ploutos.TeamUpSportGameTypes {
			if t == gameType {
				gameTypeIds = ploutos.TeamUpSportGameTypes
				teamupType = int(ploutos.TeamupTypeSports)
			}
		}

		for _, t := range ploutos.TeamUpGameGameTypes {
			if t == gameType {
				gameTypeIds = ploutos.TeamUpGameGameTypes
				teamupType = int(ploutos.TeamupTypeGames)
			}
		}

		for _, t := range ploutos.TeamUpSpinGameTypes {
			if t == gameType {
				gameTypeIds = ploutos.TeamUpSpinGameTypes
				teamupType = int(ploutos.TeamupTypeSpin)
			}
		}

		return
	}

	return
}

func GetTeamUpGameTypeSliceByBrand(brand int) (sports, games, spins []int) {

	if brand != int(BrandIdAha) && brand != int(BrandIdBatace) {
		brand = int(BrandIdBatace)
	}

	switch brand {
	case int(BrandIdBatace):
		sports = ploutos.TeamUpBASportGameTypes
		games = ploutos.TeamUpBAGameGameTypes
		spins = ploutos.TeamUpBASpinGameTypes
	case int(BrandIdAha):
		sports = ploutos.TeamUpSportGameTypes
		games = ploutos.TeamUpGameGameTypes
		spins = ploutos.TeamUpSpinGameTypes
	}

	return
}
