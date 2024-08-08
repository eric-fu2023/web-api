package service

import (
	"context"
	"errors"
	"web-api/model"
	"web-api/serializer"

	"web-api/service/common"

	"github.com/gin-gonic/gin"
)

type AnalystService struct {
	common.Page
	SportId int64 `json:"sport_id" form:"sport_id"`
}

type FollowToggle struct {
	AnalystId int64 `json:"analyst_id" form:"analyst_id"`
	// IsFollowing bool  `json:"is_following" form:"is_following"`
}

type IAnalystService interface {
	GetList(ctx context.Context) (r serializer.Response, err error)
}

func (p AnalystService) GetAnalystList(c *gin.Context) (r serializer.Response, err error) {
	// now := time.Now()
	// brand := c.MustGet(`_brand`).(int)
	// deviceInfo, _ := util.GetDeviceInfo(c)

	// analysts, err = model.AnalystList(c, p.Page, p.Limit)
	// if err != nil {
	// 	r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
	// 	return
	// }
	// r.Data = serializer.BuildAnalystList(analysts)

	data, err := model.Analyst{}.List(p.Page.Page, p.Limit, p.SportId)

	if err != nil {
		r = serializer.DBErr(c, p, "", err)
		return
	}

	r.Data = serializer.BuildAnalysts(data)

	return
}

func (service FollowToggle) FollowAnalystToggle(c *gin.Context) (r serializer.Response, err error) {

	u, _ := c.Get("user")

	user := model.User{}
	if u != nil {
		user = u.(model.User)
	}

	exist, err := model.AnalystExist(service.AnalystId)

	if err != nil {
		r = serializer.Err(c, "analyst", serializer.CodeGeneralError, "", err)
		return
	}

	if !exist {
		r = serializer.Err(c, "analyst", serializer.CodeGeneralError, "", errors.New("analyst does not exist"))
		return
	}

	following, err := model.GetFollowingAnalystStatus(c, user.ID, service.AnalystId)
	if err != nil {
		r = serializer.Err(c, "analyst", serializer.CodeGeneralError, "", err)
		return
	}

	if following.ID == 0 {
		following.UserId = user.ID
		following.AnalystId = service.AnalystId

		err = model.UpdateUserFollowAnalystStatus(following)
		if err != nil {
			r = serializer.Err(c, "analyst", serializer.CodeGeneralError, "", err)
			return
		}

		return
	}

	following.IsDeleted = !following.IsDeleted
	err = model.UpdateUserFollowAnalystStatus(following)
	if err != nil {
		r = serializer.Err(c, "analyst", serializer.CodeGeneralError, "", err)
		return
	}

	return
}

func (p AnalystService) GetFollowingAnalystList(c *gin.Context) (r serializer.Response, err error) {
	// now := time.Now()
	// brand := c.MustGet(`_brand`).(int)
	// deviceInfo, _ := util.GetDeviceInfo(c)
	u, _ := c.Get("user")

	user := model.User{}
	if u != nil {
		user = u.(model.User)
	}

	followings, err := model.GetFollowingAnalystList(c, user.ID, p.Page.Page, p.Page.Limit)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	r.Data = serializer.BuildFollowingList(followings)
	return
}

type FollowingAnalystIdsService struct{}

func (p FollowingAnalystIdsService) GetIds(c *gin.Context) (r serializer.Response, err error) {
	u, _ := c.Get("user")

	user := model.User{}
	if u != nil {
		user = u.(model.User)
	}

	followings, err := model.GetFollowingAnalystList(c, user.ID, 1, 99999)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	r.Data = serializer.BuildFollowingAnalystIdsList(followings)
	return
}

type AnalystDetailService struct {
	Id int64 `json:"analyst_id" form:"analyst_id"`
}

func (service AnalystDetailService) GetAnalyst(c *gin.Context) (r serializer.Response, err error) {
	data, err := model.Analyst{}.GetDetail(int(service.Id))

	if err != nil {
		r = serializer.DBErr(c, service, "", err)
		return
	}

	r.Data = serializer.BuildAnalystDetail(data)

	return
}

type AnalystAchievementService struct {
	AnalystId int64 `json:"analyst_id" form:"analyst_id"`
	SportId   int64 `json:"sport_id" form:"sport_id"`
}

func (service AnalystAchievementService) GetRecord(c *gin.Context) (r serializer.Response, err error) {
	// TODO : get data to fill in
	r.Data = serializer.BuildAnalystAchievement()
	return
}
