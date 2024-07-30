package service

import (
	"web-api/serializer"

	"github.com/gin-gonic/gin"
)

type TeamUpService struct {
	BetId int64 `json:"analyst_id"`
}

func (p AnalystService) ChopBet(c *gin.Context) (r serializer.Response, err error) {
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

func calculateCurrentRate(currentPercentage int, totalPercentage int, chopAmount int64) (percentage int64, err error) {

	return
}
