package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"fmt"
	"github.com/gin-gonic/gin"
	"web-api/model"
)

type Stream struct {
	ID                   int64     `json:"id"`
	Title                string    `json:"title"`
	StreamerId           int64     `json:"streamer_id"`
	MatchId              int64     `json:"match_id"`
	Status               int64     `json:"status"`
	PullUrl              string    `json:"src"`
	ImgUrl               string    `json:"img_url"`
	ScheduleTimeTS       int64     `json:"schedule_time_ts"`
	CurrentView          int64     `json:"current_view"`
	TotalView            int64     `json:"total_view"`
	StreamCategoryId     int64     `json:"category_id,omitempty"`
	StreamCategoryTypeId int64     `json:"category_type_id,omitempty"`
	RecommendStreamerId  int64     `json:"recommend_streamer_id,omitempty"`
	RoomId               string    `json:"room_id,omitempty"`
	Match                *Match    `json:"match,omitempty"`
	Streamer             *Streamer `json:"streamer,omitempty"`
}

func BuildStream(c *gin.Context, a ploutos.LiveStream) (b Stream) {
	b = Stream{
		ID:                   a.ID,
		StreamerId:           a.StreamerId,
		MatchId:              a.MatchId,
		Status:               a.Status,
		CurrentView:          a.CurrentView * 9,
		Title:                a.Title,
		PullUrl:              a.PullUrl,
		ScheduleTimeTS:       a.ScheduleTime.Unix(),
		StreamCategoryId:     a.StreamCategoryId,
		StreamCategoryTypeId: a.StreamCategoryTypeId,
	}
	if a.ImgUrl != "" {
		b.ImgUrl = Url(a.ImgUrl)
	}
	if a.StreamerId > 0 {
		b.RoomId = fmt.Sprintf(`stream:%d`, a.StreamerId)
	}
	if a.Match != nil {
		m := BuildMatch(c, *a.Match)
		b.Match = &m
	}
	if a.Streamer != nil {
		m := BuildStreamer(c, model.Streamer{
			Streamer: *a.Streamer,
			IsLive:   true,
		})
		b.Streamer = &m
	}
	return
}
