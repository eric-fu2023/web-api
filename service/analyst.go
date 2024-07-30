package service

import (
	"context"
	"web-api/serializer"

	repo "web-api/repository"
	"web-api/service/common"

	"github.com/gin-gonic/gin"
)

type AnalystService struct {
	common.Page
	Id int64 `json:"analyst_id"`
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
	r, err = analystRepo.GetListPagination(c, p.Page.Page, p.Limit)

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

func (p AnalystService) GetFollowingAnalystList(c *gin.Context) (r serializer.Response, err error) {
	// now := time.Now()
	// brand := c.MustGet(`_brand`).(int)
	// deviceInfo, _ := util.GetDeviceInfo(c)

	// analysts, err = model.GetFollowingAnalystList(c, p.Page, p.Limit)
	// if err != nil {
	// 	r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
	// 	return
	// }
	// r.Data = serializer.BuildAnalystList(analysts)

	analystRepo := repo.NewMockAnalystRepo()
	r, err = analystRepo.GetList(c)

	return
}
