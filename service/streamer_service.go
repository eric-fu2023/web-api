package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type StreamerService struct {
	Id              int64 `form:"id" json:"id" binding:"required_without=MatchId"`
	MatchId         int64 `form:"match_id" json:"match_id"`
	SportId         int64 `form:"sport_id" json:"sport_id"`
	RecommendCount  int64 `form:"recommend_count" json:"recommend_count"`
	IncludeUpcoming bool  `form:"include_upcoming" json:"include_upcoming"`
}

type StreamerWithRecommends struct {
	serializer.Streamer
	Recommends            []serializer.Stream `json:"recommends,omitempty"`
	RecommendedStreamerId int64               `json:"recommended_streamer_id,omitempty"`
}

func (service *StreamerService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var streamer model.Streamer
	if service.Id != 0 {
		streamer.ID = service.Id
	} else {
		var stream ploutos.LiveStream
		q := model.DB
		if service.SportId == 0 { // for batace
			q = q.Where(`match_id`, service.MatchId).Order(`live_streams.sort_factor DESC, live_streams.schedule_time`)
		} else {
			q = q.Scopes(model.StreamsByFbMatchIdSportId(service.MatchId, service.SportId))
		}
		if err = q.First(&stream).Error; err != nil {
			r = serializer.Err(c, service, serializer.CodeNoStream, i18n.T("stream_not_found"), err)
			return
		}
		streamer.ID = stream.StreamerId
	}

	if err = model.DB.Scopes(model.StreamerDefaultPreloads, model.StreamerWithLiveStream(service.IncludeUpcoming), model.StreamerWithGallery).First(&streamer).Error; err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("streamer_not_found"), err)
		return
	}

	if len(streamer.LiveStreams) > 0 {
		streamer.IsLive = true
	}
	data := StreamerWithRecommends{
		Streamer: serializer.BuildStreamer(c, streamer),
	}
	if len(streamer.LiveStreams) == 0 && service.RecommendCount > 0 {
		if e := model.DB.Model(model.Stream{}).Select(`recommend_streamer_id`).
			Where(`streamer_id`, streamer.ID).Where(`status`, 4).
			Order(`schedule_time DESC`).Limit(1).Find(&data.RecommendedStreamerId).Error; e == nil {
			var categories, categoryTypes []int
			if len(streamer.StreamerCategories) > 0 {
				for _, sc := range streamer.StreamerCategories {
					categories = append(categories, int(sc.CategoryId))
					categoryTypeId := int(sc.CategoryTypeId)
					var existing bool
					for _, ct := range categoryTypes {
						if ct == categoryTypeId {
							existing = true
							break
						}
					}
					if !existing {
						categoryTypes = append(categoryTypes, categoryTypeId)
					}
				}
			}
			s := StreamService{
				IncludeUpcoming:   service.IncludeUpcoming,
				CategoryOrder:     categories,
				CategoryTypeOrder: categoryTypes,
			}
			s.Page.Page = 1
			s.Page.Limit = int(service.RecommendCount)

			if list, e := s.list(c); e == nil {
				data.Recommends = list
			}
		}
	}

	r = serializer.Response{
		Data: data,
	}
	return
}
