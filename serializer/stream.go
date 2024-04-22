package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"web-api/model"
)

type Stream struct {
	ID                   int64                  `json:"id"`
	Title                string                 `json:"title"`
	StreamerId           int64                  `json:"streamer_id"`
	MatchId              int64                  `json:"match_id"`
	Status               int64                  `json:"status"`
	PullUrl              map[string]interface{} `json:"src"`
	ImgUrl               string                 `json:"img_url"`
	ScheduleTimeTS       int64                  `json:"schedule_time_ts"`
	OnlineAtTs           int64                  `json:"online_at_ts"`
	CurrentView          int64                  `json:"current_view"`
	TotalView            int64                  `json:"total_view"`
	StreamCategoryId     int64                  `json:"category_id,omitempty"`
	StreamCategoryTypeId int64                  `json:"category_type_id,omitempty"`
	RecommendStreamerId  int64                  `json:"recommend_streamer_id,omitempty"`
	ChatRoom             string                 `json:"chat_room,omitempty"`
	Width                int64                  `json:"width,omitempty"`
	Height               int64                  `json:"height,omitempty"`
	MgRoomId             *string                `json:"mg_room_id,omitempty"`
	Match                *Match                 `json:"match,omitempty"`
	Streamer             *Streamer              `json:"streamer,omitempty"`
}

func BuildStream(c *gin.Context, a ploutos.LiveStream) (b Stream) {
	b = Stream{
		ID:                   a.ID,
		StreamerId:           a.StreamerId,
		MatchId:              a.MatchId,
		Status:               a.Status,
		CurrentView:          markupNumber() + a.CurrentView*9,
		Title:                a.Title,
		ScheduleTimeTS:       a.ScheduleTime.Unix(),
		OnlineAtTs:           a.OnlineAt.Unix(),
		StreamCategoryId:     a.StreamCategoryId,
		StreamCategoryTypeId: a.StreamCategoryTypeId,
		Width:                a.Width,
		Height:               a.Height,
		MgRoomId:             a.MgRoomId,
	}
	if a.ImgUrl != "" {
		b.ImgUrl = Url(a.ImgUrl)
	}
	if a.PullUrl != "" {
		var m map[string]interface{}
		if e := json.Unmarshal([]byte(a.PullUrl), &m); e == nil {
			b.PullUrl = m
		}
	}
	if a.StreamerId > 0 {
		b.ChatRoom = fmt.Sprintf(`stream:%d`, a.ID)
	}
	if a.Match != nil {
		m := BuildMatch(c, *a.Match)
		b.Match = &m
	}
	if a.Streamer != nil {
		m := BuildStreamer(c, model.Streamer{
			User:   *a.Streamer,
			IsLive: true,
		})
		b.Streamer = &m
	}
	return
}

func markupNumber() int64 {
	return int64(rand.Intn(50) + 100)
}
