package task

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"web-api/model"
	"web-api/util"
)

func CalculateSortFactor() {
	var streams []ploutos.LiveStream
	if e := model.DB.Model(ploutos.LiveStream{}).Scopes(model.StreamsOnline).Preload(`Streamer`).Find(&streams).Error; e != nil {
		util.Log().Error("sort factor calculation error: ", e)
		return
	}
	for _, stream := range streams {
		total := stream.Streamer.SortFactor + stream.CurrentView + stream.Streamer.Followers
		if total == stream.SortFactor {
			continue
		}
		if e := model.DB.Model(ploutos.LiveStream{}).Where(`id`, stream.ID).Update(`sort_factor`, total).Error; e != nil {
			util.Log().Error("sort factor update error: ", e)
		}
	}
}
