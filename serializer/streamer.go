package serializer

import (
	"github.com/gin-gonic/gin"
	"web-api/model"
)

type Streamer struct {
	ID           int64  `json:"id"`
	Nickname     string `json:"nickname"`
	Followers    int64  `json:"follower_count"`
	Avatar       string `json:"avatar"`
	CoverImage   string `json:"cover_image"`
	Streams      int64  `json:"stream_count"`
	IsLive       bool   `json:"is_live"`
	LastLiveAtTs int64  `json:"last_live_at_ts,omitempty"`
}

func BuildStreamer(c *gin.Context, a model.Streamer) (b Streamer) {
	b = Streamer{
		ID:        a.ID,
		Nickname:  a.Nickname,
		Followers: a.Followers,
		Streams:   a.Streams,
		IsLive:    a.IsLive,
	}
	if a.Avatar != "" {
		b.Avatar = Url(a.Avatar)
	}
	if a.CoverImage != "" {
		b.CoverImage = Url(a.CoverImage)
	}
	if !a.LastLiveAt.IsZero() {
		b.LastLiveAtTs = a.LastLiveAt.Unix()
	}
	return
}
