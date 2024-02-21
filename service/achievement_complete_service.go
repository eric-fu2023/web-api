package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

var (
	// This API is meant for achievements whose progress is tracked by FE
	isAchievementWhitelisted = map[int64]bool{
		model.UserAchievementIdFirstAppLoginTutorial:     true,
		model.UserAchievementIdFirstDepositBonusTutorial: true,
	}
)

type AchievementCompleteService struct {
	AchievementId int64 `form:"achievement_id" json:"achievement_id" binding:"required"`
}

func (service *AchievementCompleteService) Complete(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	if !isAchievementWhitelisted[service.AchievementId] {
		return serializer.ParamErr(c, service, i18n.T("achievement_not_whitelisted"), err), nil
	}

	err = model.CreateUserAchievement(user.ID, service.AchievementId)
	if err != nil && !errors.Is(err, model.ErrAchievementAlreadyCompleted) {
		util.GetLoggerEntry(c).Errorf("CreateUserAchievement error: %s", err.Error())
		r = serializer.DBErr(c, service, i18n.T("general_error"), err)
		return
	}

	return serializer.Response{
		Msg: i18n.T("success"),
	}, nil
}
