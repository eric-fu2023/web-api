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

type GetTeamupService struct {
	OrderId  string `form:"order_id" json:"order_id"`
	TeamupId int64  `form:"teamup_id" json:"teamup_id"`
	common.PageNoBinding
}

type TestDepositService struct {
	UserId        int64 `form:"user_id" json:"user_id"`
	DepositAmount int64 `form:"deposit_amount" json:"deposit_amount"`
}

type DummyTeamupsService struct {
}

func (s TeamupService) List(c *gin.Context) (r serializer.Response, err error) {
	// i18n := c.MustGet("i18n").(i18n.I18n)
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

	r.Data = parseBetReport(teamupRes)

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
	var user model.User
	if u != nil {
		user = u.(model.User)
	}

	teamupRes, err := model.GetCustomTeamUpByTeamUpId(s.TeamupId)

	outgoingRes := parseBetReport(teamupRes)
	if len(outgoingRes) > 0 {
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
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	user.Avatar = serializer.Url(user.Avatar)

	br, err := model.GetTeamUpBetReport(s.OrderId)

	if err != nil {
		r = serializer.Err(c, "", serializer.CustomTeamUpBetReportDoesNotExistError, i18n.T("teamup_br_not_exist"), err)
		return
	}
	br.ParseInfo()

	if br.UserId != user.ID {
		r = serializer.Err(c, "", serializer.CustomTeamUpError, i18n.T("teamup_error"), err)
		return
	}

	var matchTime int64
	for _, bet := range br.Bets {
		if matchTime == 0 || (matchTime != 0 && matchTime >= *bet.GetMatchTime()) {
			expiredBefore, _ := model.GetAppConfigWithCache("teamup", "teamup_event_expired_before_minutes")
			expiredBeforeMinutes, _ := strconv.Atoi(expiredBefore)
			matchTime = *bet.GetMatchTime() - (60 * int64(expiredBeforeMinutes)) // 60 seconds * num minutes
		}
	}

	nowTs := time.Now().UTC().Unix()

	if nowTs >= matchTime {
		r = serializer.Err(c, "", serializer.CustomTeamUpMatchStartedError, i18n.T("teamup_match_started"), err)
		return
	}

	var teamup ploutos.Teamup
	teamup, err = model.GetTeamUp(s.OrderId)

	var t ploutos.Teamup
	t = teamup

	if teamup.ID == 0 {

		// tayaUrl, _ := model.GetAppConfig("taya_url", "apiServerAddress")
		// commonNoAuth, openAccessServiceErr := fbService.NewOpenAccessService(tayaUrl)

		// if openAccessServiceErr != nil {
		// 	log.Printf("GET OPEN ACCESS SERVICE URL=%v openAccessServiceErr err - %v \n", tayaUrl, err)
		// 	return
		// }

		var leagueIcon, homeIcon, awayIcon, leagueName string

		if len(br.Bets) > 0 {
			switch {
			case br.GameType == ploutos.GAME_FB || br.GameType == ploutos.GAME_TAYA:
				_, ok := br.Bets[0].(ploutos.BetFb)
				if ok {
					matchId, _ := strconv.Atoi(br.Bets[0].(ploutos.BetFb).GetMatchId())
					// matchDetail, err := commonNoAuth.GetMatchDetail(int64(matchId), fbServiceApi.LanguageCHINESE)
					match, err := model.GetFbMatchDetails(int64(matchId))

					if err != nil {
						log.Printf("GET MATCH DETAIL HUUUUUUUUU DEBUG FROM TAYA URL=%v commonNoAuth.GetMatchDetail err - %v \n", match, err)
					} else {
						log.Printf("GET MATCH DETAIL FROM TAYA SUCCESS %v \n", match)
						leagueIcon = match.Lg.Lurl
						leagueName = match.Lg.Na

						if len(match.Ts) > 1 {
							homeIcon = match.Ts[0].Lurl
							awayIcon = match.Ts[1].Lurl
						}
					}
				}
			case br.GameType == ploutos.GAME_IMSB:
				_, ok := br.Bets[0].(ploutos.BetImsb)

				if ok {
					// matchId := 58131174
					matchId, _ := strconv.Atoi(br.Bets[0].(ploutos.BetImsb).GetEventId())
					matches, err := model.GetImsbMatchDetails(int64(matchId))
					if err == nil && len(matches) > 0 {
						matchDetail := matches[0]
						leagueIcon = matchDetail.Competition.Format
						leagueName = matchDetail.Competition.Title

						homeIcon = matchDetail.TeamA.LogoUrl
						awayIcon = matchDetail.TeamB.LogoUrl
					}

				}
			}
		}

		teamup.UserId = user.ID
		teamup.OrderId = s.OrderId
		teamup.TotalTeamUpTarget = br.Bet
		teamup.TeamupEndTime = matchTime
		teamup.TeamupCompletedTime = matchTime

		teamup.LeagueIcon = leagueIcon
		teamup.HomeIcon = homeIcon
		teamup.AwayIcon = awayIcon
		teamup.LeagueName = leagueName

		// Recheck instead of lock
		// FUTURE: mutex for same orderId
		latestTeamup, _ := model.GetTeamUp(s.OrderId)
		if latestTeamup.ID == 0 {
			t, _ = model.SaveTeamup(teamup)
		}

	}

	shareService, err := buildTeamupShareParamsService(serializer.BuildCustomTeamupHash(t, user, br))
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
	teamup, isTeamupSuccess, isSuccess, err := model.CreateSlashBetRecord(s.TeamupId, user.ID, i18n)

	if isTeamupSuccess {

		// SUCCESS -> UPDATE USER SUM

		err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
			amount := teamup.TotalTeamUpTarget
			sum, err := model.UserSum{}.UpdateUserSumWithDB(tx, teamup.UserId, amount, amount, 0, ploutos.TransactionTypeTeamupPromotion, "")
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
		notificationMsg := fmt.Sprintf(i18n.T("notification_slashed_teamup_success"), teamup.OrderId, fmt.Sprintf("%.2f", float64(teamup.TotalTeamUpTarget)/float64(100)))
		go common.SendNotification(teamup.UserId, consts.Notification_Type_Cash_Transaction, i18n.T("notification_teamup_title"), notificationMsg)
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

func parseBetReport(teamupRes model.TeamupCustomRes) (res model.OutgoingTeamupCustomRes) {

	copier.Copy(&res, teamupRes)

	for i, t := range res {

		br := ploutos.BetReport{
			GameType: t.GameType,
			InfoJson: t.InfoJson,
		}
		br.ParseInfo()

		res[i].TotalTeamupDeposit = res[i].TotalTeamupDeposit / 100
		res[i].TotalTeamupTarget = res[i].TotalTeamupTarget / 100

		res[i].InfoJson = nil

		var outgoingBet model.OutgoingBet

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
					teams := strings.Split(outgoingBet.MatchName, " vs. ")
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

					if res[i].IsParlay {
						outgoingBet.MatchName = res[i].BetType
					}

					outgoingBet.MatchTime = teamupEndTime
					res[i].Bet = outgoingBet
				}
			}
		}
	}

	return
}

func (s TestDepositService) TestDeposit(c *gin.Context) (r serializer.Response, err error) {

	slashMultiplierString, _ := model.GetAppConfigWithCache("teamup", "teamup_slash_multiplier")
	slashMultiplier, _ := strconv.Atoi(slashMultiplierString)

	// Convert cash amount into slash progress by dividing multiplier
	contributedSlashProgress := s.DepositAmount / int64(slashMultiplier)

	err = model.GetTeamupProgressToUpdate(s.UserId, s.DepositAmount, contributedSlashProgress)
	if err != nil {
		log.Print(err.Error())
	}

	return
}
