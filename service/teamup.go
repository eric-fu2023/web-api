package service

import (
	"encoding/json"
	"math"
	"strings"
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
	common.Page
}

func (s TeamupService) List(c *gin.Context) (r serializer.Response, err error) {
	// i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

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

	teamupRes, err := model.GetAllTeamUps(user.ID, teamupStatus, s.Page.Page, s.Limit)

	for i, t := range teamupRes {

		br := ploutos.BetReport{
			GameType: t.GameType,
			InfoJson: t.InfoJson,
		}
		br.ParseInfo()

		teamupRes[i].TotalTeamupDeposit = teamupRes[i].TotalTeamupDeposit / 100
		teamupRes[i].TotalTeamupTarget = teamupRes[i].TotalTeamupTarget / 100

		teamupRes[i].TeamupProgress = (math.Min(teamupRes[i].TotalTeamupDeposit, teamupRes[i].TotalTeamupTarget) / teamupRes[i].TotalTeamupTarget) * 100

		teamupRes[i].InfoJson = nil

		if br.GameType == ploutos.GAME_FB || br.GameType == ploutos.GAME_TAYA || br.GameType == ploutos.GAME_DB_SPORT {
			var betList []model.OutgoingBet

			for _, bet := range br.Bets {
				outgoingBet := model.OutgoingBet{}
				copier.Copy(&outgoingBet, bet)
				teams := strings.Split(outgoingBet.MatchName, " vs. ")
				if len(teams) > 1 {
					outgoingBet.HomeName = teams[0]
					outgoingBet.AwayName = teams[1]
				}
				outgoingBet.HomeIcon = "https://upload.wikimedia.org/wikipedia/commons/6/66/Flag_of_Malaysia.svg"
				outgoingBet.AwayIcon = "https://icons.iconarchive.com/icons/giannis-zographos/spanish-football-club/256/Real-Madrid-icon.png"
				betList = append(betList, outgoingBet)
			}

			teamupRes[i].BetList = betList
		}

	}

	r.Data = teamupRes

	return
}

func (s GetTeamupService) Get(c *gin.Context) (r serializer.Response, err error) {

	return
}

func (s GetTeamupService) StartTeamUp(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	_, err = model.GetTeamUpBetReport(s.OrderId)

	if err != nil {
		r = serializer.DBErr(c, "", i18n.T("teamup_error"), err)
		return
	}

	var teamup ploutos.Teamup
	teamup, err = model.GetTeamUp(s.OrderId)

	if teamup.ID == 0 {
		teamup.UserId = user.ID
		teamup.OrderId = s.OrderId
		teamup.TotalTeamUpTarget = 10 * 100

		err = model.SaveTeamup(teamup)
	}

	if err != nil {
		r = serializer.DBErr(c, "", i18n.T("teamup_error"), err)
		return
	}

	shareService, err := buildTeamupShareParamsService(serializer.BuildTeamup(teamup))
	if err != nil {
		return
	}

	r, err = shareService.Create()

	return
}

func (s GetTeamupService) ContributedUserList(c *gin.Context) (r serializer.Response, err error) {
	entries, err := model.GetAllTeamUpEntries(s.TeamupId, s.Page.Page, s.Limit)

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

func (s TeamupService) SlashBet(c *gin.Context) (r serializer.Response, err error) {

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

func buildTeamupShareParamsService(teamup serializer.Teamup) (res CreateShareService, err error) {

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
