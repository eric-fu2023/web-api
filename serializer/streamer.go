package serializer

import (
	"web-api/model"

	"github.com/gin-gonic/gin"
)

type Streamer struct {
	ID                int64             `json:"id"`
	Nickname          string            `json:"nickname"`
	Followers         int64             `json:"follower_count"`
	Avatar            string            `json:"avatar"`
	CoverImage        string            `json:"cover_image"`
	Streams           int64             `json:"stream_count"`
	IsLive            bool              `json:"is_live"`
	LastLiveAtTs      int64             `json:"last_live_at_ts,omitempty"`
	AgoraUid          *int64            `json:"agora_uid,omitempty"`
	LiveStream        *Stream           `json:"live,omitempty"`
	StreamerGalleries []StreamerGallery `json:"gallery,omitempty"`
	Tags              []UserTag         `json:"tags,omitempty"`
	HasJackpot        bool              `json:"has_jackpot"`
	HasDice           bool              `json:"has_dice"`
}

func BuildStreamer(c *gin.Context, a model.Streamer) (b Streamer) {
	b = Streamer{
		ID:         a.ID,
		Nickname:   a.Nickname,
		Followers:  a.Followers,
		Streams:    a.Streams,
		IsLive:     a.IsLive,
		Avatar:     Url(a.Avatar),
		CoverImage: Url(a.CoverImage),
	}
	if !a.LastLiveAt.IsZero() {
		b.LastLiveAtTs = a.LastLiveAt.Unix()
	}
	if len(a.LiveStreams) > 0 {
		for _, s := range a.LiveStreams {
			t := BuildStream(c, s)
			b.LiveStream = &t
		}
	}
	if len(a.StreamerGalleries) > 0 {
		var galleries []StreamerGallery
		for _, g := range a.StreamerGalleries {
			galleries = append(galleries, BuildStreamerGallery(g))
		}
		b.StreamerGalleries = galleries
	}
	if len(a.UserTags) > 0 {
		for _, t := range a.UserTags {
			b.Tags = append(b.Tags, BuildUserTag(t))
		}
	}
	if a.User.UserAgoraInfo != nil && a.User.UserAgoraInfo.Uid != 0 {
		b.AgoraUid = &a.User.UserAgoraInfo.Uid
	}
	return
}
