package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	// fbService "blgit.rfdev.tech/taya/game-service/fb2/client/"
	fbService "blgit.rfdev.tech/taya/game-service/fb2/client"
	fbServiceApi "blgit.rfdev.tech/taya/game-service/fb2/client/api"

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
	r.Data = parseBetReport(teamupRes)

	return
}

func (s DummyTeamupsService) OtherTeamupList(c *gin.Context) (r serializer.Response, err error) {
	// Generate a random number between 1 and 8
	nicknameSlice := make([]string, rand.Intn(8)+1)
	for i := 0; i < len(nicknameSlice); i++ {
		nicknameSlice[i] = GetRandNickname()
	}

	r.Data = serializer.GenerateOtherTeamups(nicknameSlice)

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

	if err != nil || nowTs >= matchTime {
		r = serializer.Err(c, "", serializer.CustomTeamUpMatchStartedError, i18n.T("teamup_match_started"), err)
		return
	}

	var teamup ploutos.Teamup
	teamup, err = model.GetTeamUp(s.OrderId)

	var t ploutos.Teamup
	t = teamup

	if teamup.ID == 0 {

		tayaUrl, _ := model.GetAppConfigWithCache("taya_url", "apiServerAddress")
		commonNoAuth, openAccessServiceErr := fbService.NewOpenAccessService(tayaUrl)

		if openAccessServiceErr != nil {
			return
		}

		var leagueIcon, homeIcon, awayIcon, leagueName string

		if len(br.Bets) > 0 {
			_, ok := br.Bets[0].(ploutos.BetFb)
			matchId, _ := strconv.Atoi(br.Bets[0].(ploutos.BetFb).GetMatchId())
			if ok {
				matchDetail, err := commonNoAuth.GetMatchDetail(int64(matchId), fbServiceApi.LanguageCHINESE)

				if err == nil {
					fmt.Print(matchDetail)
					leagueIcon = matchDetail.Data.League.LeagueIconUrl
					leagueName = matchDetail.Data.League.Name

					if len(matchDetail.Data.Teams) > 1 {
						homeIcon = matchDetail.Data.Teams[0].LogoUrl
						awayIcon = matchDetail.Data.Teams[1].LogoUrl
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

		t, _ = model.SaveTeamup(teamup)
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
	isSuccess, err := model.CreateSlashBetRecord(s.TeamupId, user.ID)

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

		if br.GameType == ploutos.GAME_FB || br.GameType == ploutos.GAME_TAYA || br.GameType == ploutos.GAME_DB_SPORT {
			var outgoingBet model.OutgoingBet

			var matchTime int64

			for _, bet := range br.Bets {
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

	err = updateTeamupProgress(s.UserId, s.DepositAmount, contributedSlashProgress)
	if err != nil {
		log.Print(err.Error())
	}

	return
}

func updateTeamupProgress(userId, amount, slashProgress int64) (err error) {
	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		teamupEntry, err := model.FindOngoingTeamupEntriesByUserId(userId)
		if err != nil {
			err = fmt.Errorf("fail to get teamup err - %v", err)
			return
		}

		err = model.UpdateFirstTeamupEntryProgress(tx, teamupEntry.ID, amount, slashProgress)

		if err != nil {
			err = fmt.Errorf("fail to update teamup entry err - %v", err)
			return
		}

		err = model.UpdateTeamupProgress(tx, teamupEntry.TeamupId, amount, slashProgress)

		if err != nil {
			err = fmt.Errorf("fail to update teamup err - %v", err)
			return
		}
		return
	})

	return
}
