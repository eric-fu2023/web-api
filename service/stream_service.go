package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"
)

type StreamService struct {
	Category            string  `form:"category" json:"category"`
	CategoryOrder       []int   // only for streamer recommends use
	CategoryTypeOrder   []int   // only for streamer recommends use
	ExcludedStreamerIds []int64 // only for streamer recommends use
	common.Page
}

func (service *StreamService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var list []serializer.Stream
	if list, err = service.list(c); err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	r = serializer.Response{
		Data: list,
	}
	return
}

func (service *StreamService) list(c *gin.Context) (r []serializer.Stream, err error) {
	var categories []int
	cats := strings.Split(service.Category, ",")
	for _, c := range cats {
		if i, e := strconv.Atoi(c); e == nil {
			categories = append(categories, i)
		}
	}

	var categoryOrder string
	if len(service.CategoryOrder) > 0 {
		categoryOrder += "("
		for _, c := range service.CategoryOrder {
			categoryOrder += strconv.Itoa(c) + `,`
		}
		categoryOrder = categoryOrder[:len(categoryOrder)-1] + `)`
	}

	var categoryTypeOrder string
	if len(service.CategoryTypeOrder) > 0 {
		categoryTypeOrder += "("
		for _, c := range service.CategoryTypeOrder {
			categoryTypeOrder += strconv.Itoa(c) + `,`
		}
		categoryTypeOrder = categoryTypeOrder[:len(categoryTypeOrder)-1] + `)`
	}

	var streams []ploutos.LiveStream
	q := model.DB.Scopes(model.StreamsOnlineSorted(categoryOrder, categoryTypeOrder), model.ExcludeStreamers(service.ExcludedStreamerIds), model.Paginate(service.Page.Page, service.Limit)).Preload(`Streamer`)
	if len(categories) > 0 {
		q = q.Where(`stream_category_id`, categories)
	}
	if err = q.Find(&streams).Error; err != nil {
		return
	}
	for _, stream := range streams {
		r = append(r, serializer.BuildStream(c, stream))
	}
	return
}
