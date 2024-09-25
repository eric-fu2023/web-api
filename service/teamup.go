package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"

	"github.com/chenyahui/gin-cache/persist"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	// fbService "blgit.rfdev.tech/taya/game-service/fb2/client/"

	"github.com/gin-gonic/gin"
)

type TeamupService struct {
	Status int    `form:"status" json:"status"` // 0 = All, 1 = Ongoing, 2 = Ended
	Start  string `form:"start" json:"start" binding:"required"`
	End    string `form:"end" json:"end" binding:"required"`
	common.Page
}

type TeamupCheckSpinService struct {
	SpinId int64 `json:"spin_id"`
}

type GetTeamupService struct {
	OrderId    string `form:"order_id" json:"order_id"`
	TeamupId   int64  `form:"teamup_id" json:"teamup_id"`
	TeamupType int64  `form:"teamup_type" json:"teamup_type"`
	MatchId    string `form:"match_id" json:"match_id"`
	MatchTitle string `form:"match_title" json:"match_title"`
	IsParlay   bool   `form:"is_parlay" json:"is_parlay"`
	MarketName string `form:"market_name" json:"market_name"`
	OptionName string `form:"option_name" json:"option_name"`
	GameType   string `form:"game_type" json:"game_type"`
	MatchTime  int64  `form:"match_time" json:"match_time"`
	BetAmount  int64  `form:"bet_amount" json:"bet_amount"`
	common.PageNoBinding
}

type TestDepositService struct {
	UserId        int64 `form:"user_id" json:"user_id"`
	DepositAmount int64 `form:"deposit_amount" json:"deposit_amount"`
}

type DummyTeamupsService struct {
}

type TeamupNotificationResp struct {
	TeamupId int64 `json:"teamup_id"`
}

type TeamupGamePopUpNotification struct {
	Amount       int64  `json:"amount"`
	Icon         string `json:"icon"`
	ProviderName string `json:"provider_name"`
	EndTime      int64  `json:"end_time"`
	TeamupId     int64  `json:"teamup_id"`
	OrderId      string `json:"order_id"`
	TeamupType   int64  `json:"teamup_type"`
}

type TeamupEntrySpinResp struct {
	HasSpin bool            `json:"has_spin"`
	Spin    serializer.Spin `json:"spin"`
}

type TeamupEntrySpinResultResp struct {
	IsSuccess       bool  `json:"is_success"`
	ID              int64 `json:"id"`
	RemainingCounts int   `json:"remaining_counts"`
	TeamupId        int64 `json:"teamup_id"`
	TeamupType      int64 `json:"teamup_type"`
}

func (s TeamupService) List(c *gin.Context) (r serializer.Response, err error) {
	// i18n := c.MustGet("i18n").(i18n.I18n)
	brand := c.MustGet(`_brand`).(int)
	u, _ := c.Get("user")
	user := u.(model.User)

	var start, end int64
	loc := c.MustGet("_tz").(*time.Location)
	if s.Start != "" {
		if v, e := time.ParseInLocation(time.DateOnly, s.Start, loc); e == nil {
			start = v.UTC().Add(-10 * time.Minute).Unix()
		}
	}
	if s.End != "" {
		if v, e := time.ParseInLocation(time.DateOnly, s.End, loc); e == nil {
			end = v.UTC().Add(24*time.Hour - 1*time.Second).Add(-10 * time.Minute).Unix()
		}
	}

	teamupStatus := make([]int, 3)

	switch s.Status {

	// 进行中
	case 1:
		teamupStatus = []int{0}

	// 结束（成功，失败）
	case 2:
		teamupStatus = []int{1, 2}

	// 全部
	case 0:
		teamupStatus = []int{0, 1, 2}
	}

	teamupRes, err := model.GetAllTeamUps(user.ID, teamupStatus, s.Page.Page, s.Limit, start, end)

	// sort.SliceStable(teamupRes, func(i, j int) bool {
	// 	// Move status 0, 1, and 2 to the front
	// 	if teamupRes[i].Status == 0 || teamupRes[i].Status == 1 || teamupRes[i].Status == 2 {
	// 		if teamupRes[j].Status == 0 || teamupRes[j].Status == 1 || teamupRes[j].Status == 2 {
	// 			return teamupRes[i].Status < teamupRes[j].Status
	// 		}
	// 		return true
	// 	}
	// 	return false
	// })

	r.Data = parseBetReport(brand, teamupRes)

	return
}

func (s DummyTeamupsService) OtherTeamupList(c *gin.Context) (r serializer.Response, err error) {
	// Generate a random number between 1 and 8
	nicknameSlice := make([]string, rand.Intn(8)+1)
	for i := 0; i < len(nicknameSlice); i++ {
		nicknameSlice[i] = GetRandNickname()
	}

	successTeamups, _ := model.GetRecentCompletedSuccessTeamup(30)

	var otherTeamups []serializer.OtherTeamupContribution

	err = cache.RedisStore.Get("otherteamuplist", &otherTeamups)
	cacheRealUserCount := 0
	for _, otherTeamup := range otherTeamups {
		if otherTeamup.IsReal {
			cacheRealUserCount++
		}
	}

	if err != nil && errors.Is(err, persist.ErrCacheMiss) || len(successTeamups) != cacheRealUserCount {
		otherTeamups = serializer.GenerateOtherTeamups(nicknameSlice, successTeamups)
		err = cache.RedisStore.Set("otherteamuplist", otherTeamups, 30*time.Minute)
	}
	r.Data = otherTeamups

	return
}

func (s GetTeamupService) Get(c *gin.Context) (r serializer.Response, err error) {
	u, _ := c.Get("user")
	brand := c.MustGet(`_brand`).(int)
	var user model.User
	if u != nil {
		user = u.(model.User)
	}
	loc := c.MustGet("_tz").(*time.Location)
	log.Printf("GET MATCH DETAIL CHECK TIMEZONE tz - %v , teamup_id - %v \n", loc.String(), s.TeamupId)
	if loc.String() == "UTC" {
		loc, _ = time.LoadLocation("Asia/Tokyo")
	}
	teamupRes, err := model.GetCustomTeamUpByTeamUpId(s.TeamupId)

	outgoingRes := parseBetReport(brand, teamupRes)
	if len(outgoingRes) > 0 {
		if outgoingRes[0].TeamupEndTime != 0 && loc != nil {
			t := time.Unix(outgoingRes[0].TeamupEndTime, 0).UTC()
			tInLoc := t.In(loc)
			outgoingRes[0].TeamupLocalEndTime = tInLoc.Format("2006-01-02 15:04:05")
		}

		teamupId, _ := strconv.Atoi(outgoingRes[0].TeamupId)
		if user.ID != 0 && outgoingRes[0].UserId != fmt.Sprint(user.ID) {
			res, _ := model.GetTeamupEntryByTeamupIdAndUserId(int64(teamupId), user.ID)
			if res.ID != 0 {
				outgoingRes[0].HasJoined = true
			}
		}
		r.Data = outgoingRes[0]
	}

	return
}

func (s GetTeamupService) StartTeamUp(c *gin.Context) (r serializer.Response, err error) {

	loc := c.MustGet("_tz").(*time.Location)
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)
	brand := c.MustGet(`_brand`).(int)

	if s.GameType == "0" || s.GameType == "" {
		s.GameType = "4"
	}
	incomingGameType, _ := strconv.Atoi(s.GameType)
	_, teamupType := model.GetGameTypeSlice(brand, incomingGameType)

	if teamupType == 0 {
		r = serializer.Err(c, "", serializer.CustomTeamUpError, i18n.T("teamup_error"), err)
		return
	}

	user.Avatar = serializer.Url(user.Avatar)

	if s.TeamupId != 0 {
		if ploutos.TeamupTypeGames == s.TeamupType {
			shareService, _ := buildTeamupShareParamsService(serializer.BuildCustomTeamupGameHash(s.TeamupId, user))

			r, err = shareService.Create()

			return
		}
	}

	if s.OrderId == "" {
		r = serializer.Err(c, "", serializer.CustomTeamUpError, i18n.T("teamup_error"), err)
		return
	}

	var teamup ploutos.Teamup
	teamup, err = model.GetTeamUp(s.OrderId)

	var t ploutos.Teamup
	t = teamup

	nowTs := time.Now().UTC().Unix()

	var leagueIcon, homeIcon, awayIcon, leagueName, matchTitle, homeName, awayName string

	if teamup.ID == 0 {

		// THIS CONDITION ONLY WHEN FRONTEND ONLY PASS ORDER_ID, ASSUME ORDER_ID CAN BE USED TO GET DATA DETAILS FROM BET REPORT
		if s.MatchId == "" {
			br, _ := model.GetTeamUpBetReport(s.OrderId)

			if br.OrderId != "" {
				// if br.OrderId == "" {
				// 	r = serializer.Err(c, "", serializer.CustomTeamUpBetReportDoesNotExistError, i18n.T("teamup_br_not_exist"), err)
				// 	return
				// }
				br.ParseInfo()

				var matchTime int64

				var betMatchId, betMarketName, betOptionName, betMatchName string

				if len(br.Bets) > 0 {
					switch {
					case br.GameType == ploutos.GAME_FB || br.GameType == ploutos.GAME_TAYA:
						for _, bet := range br.Bets {
							if matchTime == 0 || (matchTime != 0 && matchTime >= *bet.GetMatchTime()) {
								betToShow := bet.(ploutos.BetFb)

								betMatchId = betToShow.MatchId
								betMarketName = betToShow.MarketName
								betOptionName = betToShow.OptionName
								betMatchName = betToShow.MatchName

								expiredBefore, _ := model.GetAppConfigWithCache("teamup", "teamup_event_expired_before_minutes")
								expiredBeforeMinutes, _ := strconv.Atoi(expiredBefore)
								matchTime = *bet.GetMatchTime() - (60 * int64(expiredBeforeMinutes)) // 60 seconds * num minutes
							}
						}
						if nowTs >= matchTime {
							r = serializer.Err(c, "", serializer.CustomTeamUpMatchStartedError, i18n.T("teamup_match_started"), err)
							return
						}

					case br.GameType == ploutos.GAME_IMSB:
						for _, bet := range br.Bets {
							if matchTime == 0 || (matchTime != 0 && matchTime >= *bet.GetMatchTime()) {
								betToShow := bet.(ploutos.BetImsb)

								betMatchId = betToShow.EventId                                // "526381"
								betMarketName = betToShow.BetType                             // "ML"
								betOptionName = betToShow.BetType + " " + betToShow.Selection // "ML H"
								betMatchName = betToShow.EventName                            // "Mi Cape Town vs Joburg Super Kings Srl"

								expiredBefore, _ := model.GetAppConfigWithCache("teamup", "teamup_event_expired_before_minutes")
								expiredBeforeMinutes, _ := strconv.Atoi(expiredBefore)
								matchTime = *bet.GetMatchTime() - (60 * int64(expiredBeforeMinutes)) // 60 seconds * num minutes
							}
						}
						if nowTs >= matchTime {
							r = serializer.Err(c, "", serializer.CustomTeamUpMatchStartedError, i18n.T("teamup_match_started"), err)
							return
						}
					}

					var teamup ploutos.Teamup
					teamup, err = model.GetTeamUp(s.OrderId)

					var t ploutos.Teamup
					t = teamup

					if len(br.Bets) > 0 {
						switch {
						case br.GameType == ploutos.GAME_FB || br.GameType == ploutos.GAME_TAYA:
							leagueIcon, leagueName, homeIcon, awayIcon, matchTitle, _, _, err = getFbMatchDetails(betMatchId)
							if err != nil {
								log.Print(err)
							}
							// if !s.IsParlay {
							// 	s.MatchTitle = homeName + " vs " + awayName
							// }
						case br.GameType == ploutos.GAME_IMSB:
							leagueIcon, leagueName, homeIcon, awayIcon, matchTitle, _, _, err = getImsbMatchDetails(betMatchId)
							if err != nil {
								log.Print(err)
							}
						}
						if !br.IsParlay {
							betMatchName = matchTitle
						}

						teamup.UserId = user.ID
						teamup.OrderId = s.OrderId
						teamup.TotalTeamUpTarget = br.Bet
						teamup.TeamupEndTime = matchTime
						teamup.TeamupCompletedTime = matchTime

						teamup.MatchTime = matchTime
						teamup.MarketName = betMarketName
						teamup.OptionName = betOptionName
						teamup.IsParlay = br.IsParlay
						teamup.MatchTitle = br.BetType
						if !br.IsParlay {
							teamup.MatchTitle = betMatchName
						}
						teamup.MatchId = betMatchId

						teamup.LeagueIcon = leagueIcon
						teamup.HomeIcon = homeIcon
						teamup.AwayIcon = awayIcon
						teamup.LeagueName = leagueName
						teamup.BetReportGameType = int(br.GameType)
						teamup.Timezone = loc.String()

						// Recheck instead of lock
						// FUTURE: mutex for same orderId
						latestTeamup, _ := model.GetTeamUp(s.OrderId)
						if latestTeamup.ID == 0 {
							t, _ = model.SaveTeamup(teamup)
						}

						shareService, _ := buildTeamupShareParamsService(serializer.BuildCustomTeamupHash(t, user, br.IsParlay))

						r, err = shareService.Create()

						return

					}

					return
				}
			} else {
				r = serializer.Err(c, "", serializer.CustomTeamUpBetReportDoesNotExistError, i18n.T("teamup_br_not_exist"), err)
			}
			return
		}

		// FRONTEND WILL GIVE EVERYTHING BECAUSE ORDER_ID HAVENT EXIST IN BET_REPORT, EVERYTHING WILL BE IN s STRUCT
		var matchExpiredTime int64

		expiredBefore, _ := model.GetAppConfigWithCache("teamup", "teamup_event_expired_before_minutes")
		expiredBeforeMinutes, _ := strconv.Atoi(expiredBefore)
		matchExpiredTime = s.MatchTime - (60 * int64(expiredBeforeMinutes)) // 60 seconds * num minutes config

		if nowTs >= matchExpiredTime {
			r = serializer.Err(c, "", serializer.CustomTeamUpMatchStartedError, i18n.T("teamup_match_started"), err)
			return
		}

		// Invalid amount if lesser than $0.01
		if s.BetAmount < 1 {
			r = serializer.Err(c, "", serializer.CustomTeamUpError, i18n.T("teamup_error"), err)
			return
		}

		/*
			IF NO TEAMUP_ID + PROVIDED MATCH_ID
			MATCH_ID PROVIDED WILL BE USED TO GET MATCH DETAILS
		*/

		switch {

		// GET MATCH DETAILS WITH FB WAY
		case s.GameType == fmt.Sprint(ploutos.GAME_FB) || s.GameType == fmt.Sprint(ploutos.GAME_TAYA):
			leagueIcon, leagueName, homeIcon, awayIcon, _, homeName, awayName, err = getFbMatchDetails(s.MatchId)
			if err != nil {
				log.Print(err)
			}

			if !s.IsParlay {
				s.MatchTitle = homeName + " vs " + awayName
			}

		// GET MATCH DETAILS WITH IMSB WAY (BatAce)
		case s.GameType == fmt.Sprint(ploutos.GAME_IMSB):
			leagueIcon, leagueName, homeIcon, awayIcon, _, homeName, awayName, err = getImsbMatchDetails(s.MatchId)
			if err != nil {
				log.Print(err)
			}

			if !s.IsParlay {
				s.MatchTitle = homeName + " vs " + awayName
			}
		}

		teamup.UserId = user.ID
		teamup.OrderId = s.OrderId
		teamup.TotalTeamUpTarget = s.BetAmount

		teamup.TeamupEndTime = matchExpiredTime
		teamup.TeamupCompletedTime = matchExpiredTime

		teamup.MatchTime = s.MatchTime
		teamup.MarketName = s.MarketName
		teamup.OptionName = s.OptionName
		teamup.IsParlay = s.IsParlay
		teamup.MatchTitle = s.MatchTitle
		teamup.MatchId = s.MatchId

		teamup.LeagueIcon = leagueIcon
		teamup.HomeIcon = homeIcon
		teamup.AwayIcon = awayIcon
		teamup.LeagueName = leagueName
		gameType, _ := strconv.Atoi(s.GameType)
		teamup.BetReportGameType = gameType
		teamup.Timezone = loc.String()

		// Recheck instead of lock
		// FUTURE: mutex for same orderId
		latestTeamup, _ := model.GetTeamUp(s.OrderId)
		if latestTeamup.ID == 0 {
			t, _ = model.SaveTeamup(teamup)
		}

	}

	shareService, err := buildTeamupShareParamsService(serializer.BuildCustomTeamupHash(t, user, s.IsParlay))
	if err != nil {
		r = serializer.Err(c, "", serializer.CustomTeamUpError, i18n.T("teamup_error"), err)
		return
	}

	r, err = shareService.Create()

	return
}

func (s GetTeamupService) ContributedUserList(c *gin.Context) (r serializer.Response, err error) {
	entries, err := model.GetAllTeamUpEntries(s.TeamupId, s.Page, s.Limit)

	for i := range entries {
		entries[i].Avatar = serializer.Url(entries[i].Avatar)
		entries[i].ContributedAmount = entries[i].ContributedAmount / float64(100)
	}

	r.Data = entries

	return
}

func (s GetTeamupService) SlashBet(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	// CREATE RECORD ONLY, THE REST WILL BE DONE IN DEPOSIT
	teamup, isTeamupSuccess, isSuccess, err := model.CreateSlashBetRecord(c, s.TeamupId, user.User, i18n)

	alreadyPushTeamupNotification := false

	if isTeamupSuccess {

		// SUCCESS -> UPDATE USER SUM

		err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
			amount := teamup.TotalTeamUpTarget
			sum, err := model.UpdateDbUserSumAndCreateTransaction(tx, teamup.UserId, amount, amount, 0, ploutos.TransactionTypeTeamupPromotion, "")
			if err != nil {
				return err
			}
			notes := fmt.Sprintf("Teamup ID - %v", teamup.ID)

			coId := uuid.NewString()
			teamupCashOrder := ploutos.CashOrder{
				ID:                    coId,
				UserId:                teamup.UserId,
				OrderType:             ploutos.CashOrderTypeTeamupPromotion,
				Status:                ploutos.CashOrderStatusSuccess,
				Notes:                 ploutos.EncryptedStr(notes),
				AppliedCashInAmount:   amount,
				ActualCashInAmount:    amount,
				EffectiveCashInAmount: amount,
				BalanceBefore:         sum.Balance - amount,
				WagerChange:           amount,
			}
			err = tx.Create(&teamupCashOrder).Error
			if err != nil {
				return err
			}

			transaction := ploutos.Transaction{
				UserId:                teamup.UserId,
				Amount:                amount,
				BalanceBefore:         sum.Balance - amount,
				BalanceAfter:          sum.Balance,
				TransactionType:       ploutos.TransactionTypeTeamupPromotion,
				Wager:                 0,
				WagerBefore:           sum.RemainingWager,
				WagerAfter:            sum.RemainingWager + amount,
				ExternalTransactionId: strconv.Itoa(int(teamup.ID)),
				GameVendorId:          0,
			}
			err = tx.Create(&transaction).Error
			if err != nil {
				log.Printf("Error creating teamup transaction, err: %v", err)
				return err
			}

			util.GetLoggerEntry(c).Info("Team Up Cash Order", coId, teamup.ID)
			common.SendUserSumSocketMsg(teamup.UserId, sum.UserSum, "teamup_success", float64(amount)/100)

			return
		})
		if err != nil {
			util.GetLoggerEntry(c).Error("Team Up Cash Order ERROR - ", err, teamup.ID)
			return
		}

		alreadyPushTeamupNotification = true
		SendTeamupNotification(2, teamup.UserId, teamup.TotalFakeProgress, teamup.TotalTeamUpTarget, teamup.ID, i18n)
	}

	if isSuccess {
		teamup, _ = model.GetTeamUpByTeamUpId(teamup.ID)
		if !alreadyPushTeamupNotification {
			SendTeamupNotification(1, teamup.UserId, teamup.TotalFakeProgress, teamup.TotalTeamUpTarget, teamup.ID, i18n)
		}
	}

	if err != nil {
		return serializer.Response{
			Code:  http.StatusBadRequest,
			Error: i18n.T(err.Error()),
		}, nil
	}

	r.Data = isSuccess

	return
}

func buildTeamupShareParamsService(teamup serializer.OutgoingTeamupHash) (res CreateShareService, err error) {

	teamupData, err := json.Marshal(teamup)
	if err != nil {
		return
	}

	jsonString := string(teamupData)

	res = CreateShareService{
		Path:   "/shareteamup",
		Params: jsonString,
	}

	return
}

func parseBetReport(brandId int, teamupRes model.TeamupCustomRes) (res model.OutgoingTeamupCustomRes) {

	copier.Copy(&res, teamupRes)

	for i, t := range teamupRes {

		var outgoingBet model.OutgoingBet

		// 游戏解析
		// 如果是游戏
		// TAKE NOTE PANDA
		_, teamupType := model.GetGameTypeSlice(brandId, t.BetReportGameType)
		res[i].TeamupType = int64(teamupType)
		res[i].TotalTeamupDeposit = res[i].TotalTeamupDeposit / 100
		res[i].TotalTeamupTarget = res[i].TotalTeamupTarget / 100

		if res[i].TeamupType != ploutos.TeamupTypeSports {
			res[i].LeagueName = consts.GameProviderNameMap[t.Provider]
			res[i].LeagueIcon = consts.GameProviderNameToImgMap[t.Provider]

			outgoingBet.LeagueName = consts.GameProviderNameMap[t.Provider]
			outgoingBet.LeagueIcon = consts.GameProviderNameToImgMap[t.Provider]

			res[i].Bet = outgoingBet

			continue
		}

		// 体育解析

		if t.MatchId == "" {
			fmt.Print(t.MatchId)
			br, _ := model.GetBetReport(t.OrderId)
			if br.OrderId != "" {
				br.ParseInfo()

				res[i].IsParlay = br.IsParlay
				res[i].BetType = br.BetType

				var matchTime int64

				for _, bet := range br.Bets {

					switch {
					case br.GameType == ploutos.GAME_FB || br.GameType == ploutos.GAME_TAYA || br.GameType == ploutos.GAME_DB_SPORT:
						if matchTime == 0 || (matchTime != 0 && matchTime >= *bet.GetMatchTime()) {
							matchTime = *bet.GetMatchTime()
							// comment out to show real match time
							// teamupEndTime := matchTime - 600 // 600 seconds = 10 minutes
							teamupEndTime := matchTime
							copier.Copy(&outgoingBet, bet)
							teams := strings.Split(outgoingBet.MatchName, " vs ")
							if len(teams) < 2 {
								teams = strings.Split(outgoingBet.MatchName, " vs. ")
							}
							if len(teams) > 1 {
								outgoingBet.HomeName = teams[0]
								outgoingBet.AwayName = teams[1]
							}
							outgoingBet.LeagueIcon = res[i].LeagueIcon
							outgoingBet.HomeIcon = res[i].HomeIcon
							outgoingBet.AwayIcon = res[i].AwayIcon

							outgoingBet.LeagueName = res[i].LeagueName

							if res[i].IsParlay {
								outgoingBet.MatchName = res[i].BetType
							}

							outgoingBet.MatchTime = teamupEndTime
							res[i].Bet = outgoingBet
						}
					case br.GameType == ploutos.GAME_IMSB:
						if matchTime == 0 || (matchTime != 0 && matchTime >= *bet.GetMatchTime()) {
							matchTime = *bet.GetMatchTime()
							// comment out to show real match time
							// teamupEndTime := matchTime - 600 // 600 seconds = 10 minutes
							teamupEndTime := matchTime
							copier.Copy(&outgoingBet, bet)
							betImsb := bet.(ploutos.BetImsb)
							teams := strings.Split(betImsb.GetMatchName(), " vs ")
							if len(teams) > 1 {
								outgoingBet.HomeName = teams[0]
								outgoingBet.AwayName = teams[1]
							}
							outgoingBet.LeagueIcon = res[i].LeagueIcon
							outgoingBet.HomeIcon = res[i].HomeIcon
							outgoingBet.AwayIcon = res[i].AwayIcon
							outgoingBet.LeagueName = betImsb.GetCompetitionName()
							outgoingBet.MarketName = betImsb.BetType
							outgoingBet.OptionName = betImsb.Selection + " " + fmt.Sprint(betImsb.Odds)
							outgoingBet.MatchName = betImsb.EventName

							if res[i].IsParlay {
								outgoingBet.MatchName = res[i].BetType
							}

							outgoingBet.MatchTime = teamupEndTime
							res[i].Bet = outgoingBet
						}
					}
				}

			}

			continue
		}

		res[i].InfoJson = nil

		// var matchTime int64

		switch {
		case t.BetReportGameType == ploutos.GAME_FB || t.BetReportGameType == ploutos.GAME_TAYA || t.BetReportGameType == ploutos.GAME_DB_SPORT:
			outgoingBet.MatchTime = t.MatchTime
			outgoingBet.HomeName = t.HomeName
			outgoingBet.HomeIcon = t.HomeIcon
			outgoingBet.AwayName = t.AwayName
			outgoingBet.AwayIcon = t.AwayIcon
			outgoingBet.LeagueIcon = t.LeagueIcon
			outgoingBet.LeagueName = t.LeagueName
			outgoingBet.MarketName = t.MarketName
			outgoingBet.MatchName = t.MatchTitle
			outgoingBet.LeagueName = t.LeagueName
			outgoingBet.OptionName = t.OptionName
			teams := strings.Split(outgoingBet.MatchName, " vs ")
			if len(teams) < 2 {
				teams = strings.Split(outgoingBet.MatchName, " vs. ")
			}
			if len(teams) > 1 {
				outgoingBet.HomeName = teams[0]
				outgoingBet.AwayName = teams[1]
			}

		case t.BetReportGameType == ploutos.GAME_IMSB:
			outgoingBet.MatchTime = t.MatchTime
			outgoingBet.HomeName = t.HomeName
			outgoingBet.HomeIcon = t.HomeIcon
			outgoingBet.AwayName = t.AwayName
			outgoingBet.AwayIcon = t.AwayIcon
			outgoingBet.LeagueIcon = t.LeagueIcon
			outgoingBet.LeagueName = t.LeagueName
			outgoingBet.MarketName = t.MarketName
			outgoingBet.MatchName = t.MatchTitle
			outgoingBet.LeagueName = t.LeagueName
			outgoingBet.OptionName = t.OptionName
			teams := strings.Split(outgoingBet.MatchName, " vs ")
			if len(teams) < 2 {
				teams = strings.Split(outgoingBet.MatchName, " vs. ")
			}
			if len(teams) < 2 {
				teams = strings.Split(outgoingBet.MatchName, " VS ")
			}
			if len(teams) < 2 {
				teams = strings.Split(outgoingBet.MatchName, " VS. ")
			}
			if len(teams) > 1 {
				outgoingBet.HomeName = teams[0]
				outgoingBet.AwayName = teams[1]
			}

		}

		res[i].Bet = outgoingBet

	}

	return
}

func getFbMatchDetails(betMatchId string) (leagueIcon, leagueName, homeIcon, awayIcon, matchTitle, homeName, awayName string, err error) {
	matchId, _ := strconv.Atoi(betMatchId)
	match, err := model.GetFbMatchDetails(int64(matchId))

	if err != nil {
		log.Printf("GET MATCH DETAIL ERROR DEBUG FROM TAYA URL=%v GetMatchDetail err - %v \n", match, err)
		err = fmt.Errorf("fail to get fb match detail err - %v", err)

		return
	}

	log.Printf("GET MATCH DETAIL FROM TAYA SUCCESS %v \n", match)
	leagueIcon = match.Lg.Lurl
	leagueName = match.Lg.Na

	if len(match.Ts) > 1 {
		homeIcon = match.Ts[0].Lurl
		awayIcon = match.Ts[1].Lurl

		homeName = match.Ts[0].Na
		awayName = match.Ts[1].Na
	}

	matchTitle = match.Nm

	return
}

func getImsbMatchDetails(betMatchId string) (leagueIcon, leagueName, homeIcon, awayIcon, matchTitle, homeName, awayName string, err error) {
	matchId, _ := strconv.Atoi(betMatchId)
	matches, _ := model.GetImsbMatchDetails(int64(matchId))
	if len(matches) == 0 {
		// Unable to get match detail
		log.Printf("GET MATCH DETAIL ERROR DEBUG FROM IMSB URL=%v GetMatchDetail err - %v \n", matches, err)
		err = fmt.Errorf("fail to get imsb match detail err - %v", err)
		return
	}

	matchDetail := matches[0]
	log.Printf("GET MATCH DETAIL FROM IMSB SUCCESS %v \n", matchDetail)

	matchTitle = matchDetail.Title
	leagueIcon = matchDetail.Competition.Format
	leagueName = matchDetail.Competition.Title
	homeIcon = matchDetail.TeamA.LogoUrl
	awayIcon = matchDetail.TeamB.LogoUrl

	homeName = matchDetail.TeamA.Name
	awayName = matchDetail.TeamB.Name

	return
}

func (s TestDepositService) TestDeposit(c *gin.Context) (r serializer.Response, err error) {

	return
}

func SendTeamupNotification(teamupType int, userId, percentage, totalTarget, teamupId int64, i18n i18n.I18n) {

	// TYPE 1 = PROGRESS
	// TYPE 2 = SUCCESS

	if teamupType > 2 || teamupType == 0 {
		teamupType = 1
	}

	n := rand.Intn(3)

	titles := []string{}
	contents := []string{}

	switch {
	case teamupType == 1:
		titles = []string{i18n.T("notification_slash_teamup_progress_title1"), i18n.T("notification_slash_teamup_progress_title2"), i18n.T("notification_slash_teamup_progress_title3")}
		contents = []string{i18n.T("notification_slash_teamup_progress_content1"), i18n.T("notification_slash_teamup_progress_content2"), i18n.T("notification_slash_teamup_progress_content3")}

	case teamupType == 2:
		titles = []string{i18n.T("notification_slash_teamup_success_title1"), i18n.T("notification_slash_teamup_success_title2"), i18n.T("notification_slash_teamup_success_title3")}
		contents = []string{i18n.T("notification_slash_teamup_success_content1"), i18n.T("notification_slash_teamup_success_content2"), i18n.T("notification_slash_teamup_success_content3")}
	}

	notificationTitle := titles[n]
	notificationMsg := contents[n]
	if strings.Contains(notificationMsg, "%s") {
		// pString := fmt.Sprintf("%.2f", (float64(percentage)/float64(10000))*(float64(totalTarget)/float64(100)))
		// notificationMsg = fmt.Sprintf(notificationMsg, pString)

		// if n != 1 {
		// 	pFloat64 := (float64(percentage) / float64(100))
		// 	notificationMsg = strings.ReplaceAll(notificationMsg, pString, "%s")
		// 	notificationMsg = fmt.Sprintf(notificationMsg, fmt.Sprintf("%.2f", pFloat64)+"%")
		// }

		pFloat64 := (float64(percentage) / float64(100))
		notificationMsg = fmt.Sprintf(notificationMsg, fmt.Sprintf("%.2f", pFloat64)+"%")
	}

	var resp serializer.Response

	resp.Data = TeamupNotificationResp{
		TeamupId: teamupId,
	}

	go common.SendNotification(userId, consts.Notification_Type_Teamup_Detail, notificationTitle, notificationMsg, resp)

}

func (s TeamupCheckSpinService) CheckSpinPopup(c *gin.Context) (r serializer.Response, err error) {
	// i18n := c.MustGet("i18n").(i18n.I18n)
	brand := c.MustGet(`_brand`).(int)
	u, _ := c.Get("user")
	user := u.(model.User)

	shouldPop := model.ShouldPopRoulette(brand, user.ID)

	spinRes := TeamupEntrySpinResp{
		HasSpin: shouldPop,
	}

	if shouldPop {
		teamUpSpinPromotionIdString, _ := model.GetAppConfigWithCache("teamup", "teamup_spin_promotion_id")

		spin, getSpinErr := model.GetSpinByPromotionId(teamUpSpinPromotionIdString)
		if getSpinErr != nil {
			err = getSpinErr
			return
		}

		spinItems, getSpinErr := model.GetSpinItemsBySpinId(spin.ID)
		if getSpinErr != nil {
			err = getSpinErr
			return
		}

		spinRes.Spin = serializer.BuildSpin(spin, spinItems, 1, true)
	}

	r.Data = spinRes

	return
}

func (s TeamupCheckSpinService) TeamupSpinResult(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)
	brand := c.MustGet(`_brand`).(int)

	shouldPop := model.ShouldPopRoulette(brand, user.ID)

	if !shouldPop {
		r.Data = TeamupEntrySpinResultResp{
			IsSuccess: false,
		}

		return
	}

	teamUpSpinPromotionIdString, _ := model.GetAppConfigWithCache("teamup", "teamup_spin_promotion_id")
	spin, err := model.GetSpinByPromotionId(teamUpSpinPromotionIdString)
	if err != nil {
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	spinItems, err := model.GetSpinItemsBySpinId(spin.ID)

	if err != nil {
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	//------------------------------------- choose the spin item based on the probability
	var resultSpinItem ploutos.SpinItem
	var totalProbability float64
	for _, item := range spinItems {
		totalProbability += item.Probability
	}
	// Normalize the probabilities so they sum to 1.
	normalizedProbabilities := make([]float64, len(spinItems))
	for i, item := range spinItems {
		normalizedProbabilities[i] = item.Probability / totalProbability
	}
	// Generate a random number between 0 and 1.
	randomValue := rand.Float64()
	// Use the random number to select an item.
	var cumulativeProbability float64
	for i, probability := range normalizedProbabilities {
		cumulativeProbability += probability
		if randomValue < cumulativeProbability {
			resultSpinItem = spinItems[i]
			break
		}
	}
	// ------------------------------------- end of choose the spin item based on the probability

	data := serializer.BuildSpinResult(resultSpinItem, 0)

	spinRes := ploutos.SpinResult{
		UserID:     user.ID,
		SpinResult: data.ID,
		Redeemed:   true,
		SpinID:     spin.ID,
	}
	err = model.DB.Create(&spinRes).Error
	if err != nil {
		fmt.Println("spin result insert err", err)
		return
	}

	var teamup ploutos.Teamup
	currTime := time.Now().UTC()

	teamup.UserId = user.ID
	teamup.OrderId = fmt.Sprint(spinRes.ID)
	teamup.TotalTeamUpTarget += int64(resultSpinItem.Amount)
	teamup.TeamupEndTime = currTime.Add(7 * 24 * time.Hour).Unix()
	teamup.TeamupCompletedTime = currTime.Add(7 * 24 * time.Hour).Unix()
	teamup.Provider = "Spin And Win Rewards"
	teamup.BetReportGameType = ploutos.TeamupSpinWheelId
	teamup.Status = int(ploutos.TeamupStatusPending)

	if resultSpinItem.Amount == 0 {
		teamup.Status = int(ploutos.TeamupStatusInvalidAmount)
	}

	t, err := model.SaveTeamup(teamup)

	if err != nil {
		fmt.Println("create teamup insert err", err)
		return
	}

	_, teamupType := model.GetGameTypeSlice(brand, t.BetReportGameType)

	r.Data = TeamupEntrySpinResultResp{
		IsSuccess:       true,
		ID:              data.ID,
		RemainingCounts: data.RemainingCounts,
		TeamupId:        t.ID,
		TeamupType:      int64(teamupType),
	}

	return
}
