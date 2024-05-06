package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"
)

type StreamService struct {
	IncludeUpcoming     bool    `form:"include_upcoming" json:"include_upcoming"`
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
	isA := false
	if v, exists := c.Get("_isA"); exists {
		if vv, ok := v.(bool); ok && vv {
			isA = vv
		}
	}
	var streams []ploutos.LiveStream
	q := model.DB.Scopes(
		model.StreamsOnlineSorted(categoryOrder, categoryTypeOrder, service.IncludeUpcoming),
		model.ExcludeStreamers(service.ExcludedStreamerIds),
		model.ExcludeStreamSource([]int64{888}),
		model.Paginate(service.Page.Page, service.Limit),
		model.StreamsABStreamSource(isA)).
		Preload(`Streamer`).
		Preload(`Streamer.UserAgoraInfo`)
	if len(categories) > 0 {
		q = q.Where(`stream_category_id`, categories)
	}
	if err = q.Find(&streams).Error; err != nil {
		return
	}
	if service.Page.Page == 1 {
		var selectedStreams []ploutos.LiveStream
		qq := model.DB.Model(ploutos.LiveStream{}).Scopes(model.StreamsABStreamSource(isA)).
			Preload(`Streamer`).Preload(`Streamer.UserAgoraInfo`).
			Where(`stream_source`, []int64{888}).Where(`status`, 2)
		if len(categories) > 0 {
			qq = qq.Where(`stream_category_id`, categories)
		}
		err = qq.Find(&selectedStreams).Error
		if err != nil {
			return
		}
		if len(selectedStreams) > 0 {
			var ind int
			if len(streams) > 0 {
				if len(streams) > 4 {
					ind = 3
				} else {
					ind = len(streams) - 1
				}
			}
			var subStreams []ploutos.LiveStream
			subStreams = append(subStreams, streams[:ind]...)
			subStreams = append(subStreams, selectedStreams...)
			subStreams = append(subStreams, streams[ind:]...)
			streams = subStreams
		}
	}
	for _, stream := range streams {
		r = append(r, serializer.BuildStream(c, stream))
	}
	return
}

type StreamStatusService struct {
	StreamerId int64 `form:"streamer_id" json:"streamer_id"`
}

func (service *StreamStatusService) Get(c *gin.Context) (r serializer.Response, err error) {
	u, _ := c.Get("user")
	user := u.(model.User)

	var silencedUntil int
	re := cache.RedisSessionClient.Get(context.TODO(), fmt.Sprintf("silenced:%d:%d", service.StreamerId, user.ID))
	if re.Err() == nil {
		if v, e := strconv.Atoi(re.Val()); e == nil {
			silencedUntil = v
		}
	}

	r = serializer.Response{
		Data: silencedUntil,
	}
	return
}
