package service

import (
	"encoding/json"
	"web-api/serializer"
	"web-api/service/common"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type TeamupService struct {
	IsEnded *bool  `form:"is_ended" json:"is_ended"`
	Start   string `form:"start" json:"start" binding:"required"`
	End     string `form:"end" json:"end" binding:"required"`
	common.Page
}

type GetTeamupService struct {
	TeamupId int64 `form:"teamup_id" json:"teamup_id"`
}

func (s TeamupService) List(c *gin.Context) (r serializer.Response, err error) {

	return
}

func (s GetTeamupService) Get(c *gin.Context) (r serializer.Response, err error) {

	return
}

func (s GetTeamupService) StartTeamUp(c *gin.Context) (r serializer.Response, err error) {

	var teamup ploutos.Teamup

	teamup.ID = 888888

	shareService, err := buildTeamupShareParamsService(serializer.BuildTeamup(teamup))
	if err != nil {
		return
	}

	r, err = shareService.Create()

	return
}

func (s TeamupService) ChopBet(c *gin.Context) (r serializer.Response, err error) {
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
