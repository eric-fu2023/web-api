package service

import (
	"encoding/json"
	"math"
	"math/rand"
	"strings"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"

	"github.com/jinzhu/copier"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

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

type DummyTeamupsService struct {
}

func (s TeamupService) List(c *gin.Context) (r serializer.Response, err error) {
	// i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var start, end time.Time
	loc := c.MustGet("_tz").(*time.Location)
	if s.Start != "" {
		if v, e := time.ParseInLocation(time.DateOnly, s.Start, loc); e == nil {
			start = v.UTC().Add(-10 * time.Minute)
		}
	}
	if s.End != "" {
		if v, e := time.ParseInLocation(time.DateOnly, s.End, loc); e == nil {
			end = v.UTC().Add(24*time.Hour - 1*time.Second).Add(-10 * time.Minute)
		}
	}

	teamupStatus := make([]int, 3)

	switch s.Status {
	case 1:
		teamupStatus = []int{0}
	case 2:
		teamupStatus = []int{1, 2}
	case 0:
	default:
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

	teamupRes, err := model.GetCustomTeamUpByTeamUpId(s.TeamupId)

	outgoingRes := parseBetReport(teamupRes)
	if len(outgoingRes) > 0 {
		r.Data = outgoingRes[0]
	}

	return
}

func (s GetTeamupService) StartTeamUp(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	br, err := model.GetTeamUpBetReport(s.OrderId)
	br.ParseInfo()

	if err != nil {
		r = serializer.DBErr(c, "", i18n.T("teamup_match_started"), err)
		return
	}

	var matchTime int64
	for _, bet := range br.Bets {
		if matchTime == 0 || (matchTime != 0 && matchTime >= *bet.GetMatchTime()) {
			matchTime = *bet.GetMatchTime() - (600) // 600 = 10 mins for timestamp
		}
	}

	nowTs := time.Now().UTC().Unix()

	if err != nil || nowTs >= matchTime {
		r = serializer.DBErr(c, "", i18n.T("teamup_match_started"), err)
		return
	}

	var teamup ploutos.Teamup
	teamup, err = model.GetTeamUp(s.OrderId)

	if teamup.ID == 0 {
		teamup.UserId = user.ID
		teamup.OrderId = s.OrderId
		teamup.TotalTeamUpTarget = 10 * 100
		teamup.TeamupEndTime = matchTime
		teamup.TeamupCompletedTime = matchTime

		err = model.SaveTeamup(teamup)
	}

	if err != nil {
		r = serializer.DBErr(c, "", i18n.T("teamup_error"), err)
		return
	}

	shareService, err := buildTeamupShareParamsService(serializer.BuildCustomTeamupHash(teamup, user, br))
	if err != nil {
		return
	}

	r, err = shareService.Create()

	return
}

func (s GetTeamupService) ContributedUserList(c *gin.Context) (r serializer.Response, err error) {
	entries, err := model.GetAllTeamUpEntries(s.TeamupId, s.Page, s.Limit)

	for i, _ := range entries {
		entries[i].ContributedAmount = entries[i].ContributedAmount / float64(100)
	}

	r.Data = entries

	// now := time.Now()
	// brand := c.MustGet(`_brand`).(int)
	// deviceInfo, _ := util.GetDeviceInfo(c)

	// analysts, err = model.AnalystList(c, p.Page, p.Limit)
	// if err != nil {
	// 	r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
	// 	return
	// }
	// r.Data = serializer.BuildAnalystList(analysts)

	// analystRepo := repo.NewMockAnalystRepo()
	// r, err = analystRepo.GetList(c)

	return
}

func (s GetTeamupService) SlashBet(c *gin.Context) (r serializer.Response, err error) {

	// now := time.Now()
	// brand := c.MustGet(`_brand`).(int)
	// deviceInfo, _ := util.GetDeviceInfo(c)
	u, _ := c.Get("user")
	user := u.(model.User)

	// analysts, err = model.AnalystList(c, p.Page, p.Limit)
	// if err != nil {
	// 	r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
	// 	return
	// }
	// r.Data = serializer.BuildAnalystList(analysts)

	// analystRepo := repo.NewMockAnalystRepo()
	// r, err = analystRepo.GetList(c)

	// CREATE RECORD ONLY, THE REST WILL BE DONE IN DEPOSIT
	isSuccess, err := model.CreateSlashBetRecord(s.TeamupId, user.ID)

	// Find user current slashed, find first not completed / expired

	// topup will add to the very first

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

		res[i].TeamupProgress = (math.Min(res[i].TotalTeamupDeposit, res[i].TotalTeamupTarget) / res[i].TotalTeamupTarget) * 100

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
					outgoingBet.LeagueIcon = "https://upload.wikimedia.org/wikipedia/commons/6/66/Flag_of_Malaysia.svg"
					outgoingBet.LeagueName = "欧美中联赛"
					outgoingBet.HomeIcon = "https://upload.wikimedia.org/wikipedia/commons/6/66/Flag_of_Malaysia.svg"
					outgoingBet.AwayIcon = "https://icons.iconarchive.com/icons/giannis-zographos/spanish-football-club/256/Real-Madrid-icon.png"

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
