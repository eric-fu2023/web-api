package stream_game

import (
	"blgit.rfdev.tech/taya/game-service/game/stream_game_api"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"
)

type GetStreamGame struct {
	Ongoing     *serializer.StreamGameSession  `json:"ongoing,omitempty"`
	ResultCount *int64                         `json:"result_count,omitempty"`
	Results     []serializer.StreamGameSession `json:"results,omitempty"`
}

type StreamGameService struct {
	GameId   int64 `form:"game_id" json:"game_id"`
	StreamId int64 `form:"stream_id" json:"stream_id" binding:"required"`
	common.PageById
}

func (service *StreamGameService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	gameService := stream_game_api.NewService(model.DB)
	cacheInfo := model.CacheInfo{
		Prefix: fmt.Sprintf(`query:stream_game:%d`, service.StreamId),
		Ttl:    10,
	}
	ctx := context.WithValue(context.TODO(), model.KeyCacheInfo, cacheInfo)
	res, err := gameService.GetGameSession(ctx, service.GameId, service.StreamId, service.PageById.IdFrom, service.PageById.Limit)
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var data *GetStreamGame
	if res.Ongoing.ID != 0 || len(res.Results) != 0 {
		var d GetStreamGame
		d.ResultCount = res.ResultCount
		if res.Ongoing.ID != 0 {
			t := serializer.BuildStreamGameSession(c, res.Ongoing)
			d.Ongoing = &t
		}
		for _, rr := range res.Results {
			t := serializer.BuildStreamGameSession(c, rr)
			d.Results = append(d.Results, t)
		}
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
	gameService := stream_game_api.NewService(model.DB)
	games, err := gameService.ListGames()
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
