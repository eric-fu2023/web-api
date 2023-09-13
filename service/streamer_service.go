package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type StreamerService struct {
	Id             int64 `form:"id" json:"id" binding:"required"`
	RecommendCount int64 `form:"recommend_count" json:"recommend_count"`
}

type StreamerWithRecommends struct {
	serializer.Streamer
	Recommends            []serializer.Stream `json:"recommends,omitempty"`
	RecommendedStreamerId int64               `json:"recommended_streamer_id,omitempty"`
}

func (service *StreamerService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var streamer ploutos.Streamer
	streamer.ID = service.Id
	if err = model.DB.Scopes(model.StreamerWithLiveStream, model.StreamerWithGallery).Preload(`StreamerCategories`).First(&streamer).Error; err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("streamer_not_found"), err)
		return
	}
	var isLive bool
	if len(streamer.LiveStreams) > 0 {
		isLive = true
	}
	data := StreamerWithRecommends{
		Streamer: serializer.BuildStreamer(c, model.Streamer{
			Streamer: streamer,
			IsLive:   isLive,
		}),
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
