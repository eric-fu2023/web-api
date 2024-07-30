package service

import (
	"context"
	"web-api/model"
	"web-api/serializer"

	repo "web-api/repository"
	"web-api/service/common"

	"github.com/gin-gonic/gin"
)

type AnalystService struct {
	common.Page
	Id int64 `json:"analyst_id"`
}

type FollowToggle struct {
	AnalystId   int64 `json:"analyst_id"`
	IsFollowing bool  `json:"is_following"`
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

	analystRepo := repo.NewMockAnalystRepo()
	r, err = analystRepo.GetList(c)

	return
}

func (service AnalystService) GetAnalyst(c *gin.Context) (r serializer.Response, err error) {
	// now := time.Now()
	// brand := c.MustGet(`_brand`).(int)
	// deviceInfo, _ := util.GetDeviceInfo(c)

	// analyst, err = model.GetAnalyst(c, service.Id)
	// if err != nil {
	// 	r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
	// 	return
	// }
	// r.Data = serializer.BuildAnalyst(analyst)

	analystRepo := repo.NewMockAnalystRepo()
	r, err = analystRepo.GetList(c)

	return
}

func (service FollowToggle) FollowAnalystToggle(c *gin.Context) (r serializer.Response, err error) {

	u, _ := c.Get("user")

	user := model.User{}
	if u != nil {
		user = u.(model.User)
	}

	// TODO: Check If Analyst Exist

	// analyst, err := model.GetAnalyst(c, service.AnalystId)
	// if analyst.Id == 0 || err != nil {
	// 	if err != nil {
	// 		r = serializer.Err(c, "analyst", serializer.CodeGeneralError, "", err)
	// 		return
	// 	}
	// }

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

	analystRepo := repo.NewMockAnalystRepo()
	r, err = analystRepo.GetList(c)

	return
}
