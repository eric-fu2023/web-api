package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type OrderListService struct {
	IsParlay  bool   `form:"is_parlay" json:"is_parlay"`
	IsSettled bool   `form:"is_settled" json:"is_settled"`
	Start     string `form:"start" json:"start" binding:"required"`
	End       string `form:"end" json:"end" binding:"required"`
}

func (service *OrderListService) List(c *gin.Context) serializer.Response {
	var err error
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var list []ploutos.BetReport
	var start, end time.Time
	loc := c.MustGet("_tz").(*time.Location)
	if service.Start != "" {
		if v, e := time.ParseInLocation(time.DateOnly, service.Start, loc); e == nil {
			start = v.UTC()
		}
	}
	if service.End != "" {
		if v, e := time.ParseInLocation(time.DateOnly, service.End, loc); e == nil {
			end = v.UTC().Add(24*time.Hour - 1*time.Second)
		}
	}
	err = model.DB.Model(ploutos.BetReport{}).Scopes(model.ByOrderListConditions(user.ID, service.IsParlay, service.IsSettled, start, end)).Find(&list).Error
	if err != nil {
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	for i, l := range list {
		l.ParseInfo()
		list[i] = l
	}

	var data []serializer.BetReport
	for _, l := range list {
		data = append(data, serializer.BuildBetReportFb(c, l))
	}

	return serializer.Response{
		Data: data,
	}
}
