package service

import (
	"encoding/json"
	"strconv"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"

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

	teamups, err := model.GetAllTeamUps(user.ID, teamupStatus, s.Page.Page, s.Limit)

	r.Data = teamups

	return
}

func (s GetTeamupService) Get(c *gin.Context) (r serializer.Response, err error) {

	return
}

func (s GetTeamupService) StartTeamUp(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var betReport ploutos.BetReport
	err = model.DB.Where("business_id = ?", s.OrderId).First(&betReport).Error

	if err != nil {
		r = serializer.DBErr(c, "", i18n.T("teamup_error"), err)
		return
	}

	var teamup ploutos.Teamup
	teamup, err = model.GetTeamUp(s.OrderId)

	if teamup.ID == 0 {
		teamup.UserId = user.ID
		orderId, _ := strconv.Atoi(s.OrderId)
		teamup.OrderId = int64(orderId)
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
