package service

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

type AchievementListService struct {
	AchievementIds string `form:"achievement_ids" json:"achievement_ids"`
}

func (service *AchievementListService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var achievementIds []int64
	for _, s := range strings.Split(service.AchievementIds, ",") {
		if i, e := strconv.Atoi(strings.TrimSpace(s)); e == nil {
			achievementIds = append(achievementIds, int64(i))
		}
	}

	uaCond := model.GetUserAchievementCond{AchievementIds: achievementIds}
	userAchievements, err := model.GetUserAchievements(user.ID, uaCond)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetUserAchievements error: %s", err.Error())
		r = serializer.DBErr(c, service, i18n.T("general_error"), err)
		return
	}

	type RespData struct {
		Achievements []serializer.UserAchievement `json:"achievements"`
	}
	respData := RespData{Achievements: serializer.BuildUserAchievements(userAchievements)}

	return serializer.Response{
		Msg:  i18n.T("success"),
		Data: respData,
	}, nil
}
