package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/json"
	"github.com/gin-gonic/gin"
)

type StreamGame struct {
	Id     int64       `json:"id"`
	Name   string      `json:"game_name"`
	Url    string      `json:"url"`
	Config interface{} `json:"config"`
}

func BuildStreamGame(c *gin.Context, a ploutos.StreamGame) (b StreamGame) {
	b = StreamGame{
		Id:   a.ID,
		Name: a.Name,
		Url:  Url(a.Url),
	}
	if a.Config != nil {
		var v map[string]interface{}
		if e := json.Unmarshal(a.Config, &v); e == nil {
			b.Config = v
		}
	}
	return
}
