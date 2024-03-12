package stream_game

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"
)

type GetStreamGame struct {
	Ongoing *serializer.StreamGameSession  `json:"ongoing,omitempty"`
	Results []serializer.StreamGameSession `json:"results,omitempty"`
}

type StreamGameService struct {
	StreamId int64 `form:"stream_id" json:"stream_id" binding:"required"`
	common.PageById
}

func (service *StreamGameService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var ongoing ploutos.StreamGameSession
	if service.PageById.IdFrom == 0 { // ongoing will only show on first page
		err = model.DB.Where(`reference_id`, service.StreamId).Where(`status`, ploutos.StreamGameSessionStatusOpen).
			Order(`created_at DESC, id DESC`).Limit(1).Find(&ongoing).Error
		if err != nil {
			r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
			return
		}
	}
	var results []ploutos.StreamGameSession
	q := model.DB.Where(`reference_id`, service.StreamId).
		Where(`status`, []int64{ploutos.StreamGameSessionStatusComplete, ploutos.StreamGameSessionStatusSettled}).
		Order(`created_at DESC, id DESC`).Limit(service.PageById.Limit)
	if service.PageById.IdFrom != 0 {
		q = q.Where(`id < ?`, service.PageById.IdFrom)
	}
	err = q.Find(&results).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var d GetStreamGame
	if ongoing.Id != 0 {
		t := serializer.BuildStreamGameSession(c, ongoing)
		d.Ongoing = &t
	}
	for _, rr := range results {
		t := serializer.BuildStreamGameSession(c, rr)
		d.Results = append(d.Results, t)
	}
	var data *GetStreamGame
	if d.Ongoing != nil || d.Results != nil {
		data = &d
	}
	r = serializer.Response{
		Data: data,
	}
	return
}

type StreamGameServiceList struct {
}

func (service *StreamGameServiceList) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var games []ploutos.StreamGame
	err = model.DB.Model(ploutos.StreamGame{}).Order(`id`).Find(&games).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var data []serializer.StreamGame
	for _, g := range games {
		data = append(data, serializer.BuildStreamGame(c, g))
	}
	r = serializer.Response{
		Data: data,
	}
	return
}
