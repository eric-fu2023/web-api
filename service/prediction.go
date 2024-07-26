package service

import (
	"context"
	"web-api/serializer"

	repo "web-api/repository"
	"web-api/service/common"

	"github.com/gin-gonic/gin"
)

type PredictionService struct {
	common.Page
}

type IPredictionService interface {
	GetList(ctx context.Context) (r serializer.Response, err error)
}

func (p PredictionService) GetList(c *gin.Context) (r serializer.Response, err error) {
	// now := time.Now()
	// brand := c.MustGet(`_brand`).(int)
	// deviceInfo, _ := util.GetDeviceInfo(c)

	// analysts, err = model.AnalystList(c, p.Page, p.Limit)
	// if err != nil {
	// 	r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
	// 	return
	// }
	// r.Data = serializer.BuildAnalystList(analysts)

	predictionRepo := repo.NewMockAnalystRepo()
	r, err = predictionRepo.GetList(c)

	return
}
