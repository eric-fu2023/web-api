package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/json"
	"github.com/gin-gonic/gin"
)

type StreamGameSession struct {
	Id          int64       `json:"id"`
	GameId      int64       `json:"game_id"`
	ReferenceId int64       `json:"reference_id"`
	Result      interface{} `json:"result,omitempty"`
	StreamGame  *StreamGame `json:"game,omitempty"`
}

func BuildStreamGameSession(c *gin.Context, a ploutos.StreamGameSession) (b StreamGameSession) {
	b = StreamGameSession{
		Id:          a.ID,
		GameId:      a.StreamGameId,
		ReferenceId: a.ReferenceId,
	}
	if a.Result != "" && a.Result != "{}" {
		var j map[string]interface{}
		if e := json.Unmarshal([]byte(a.Result), &j); e == nil {
			b.Result = j
		}
	}
	if a.StreamGame != nil {
		t := BuildStreamGame(c, *a.StreamGame)
		b.StreamGame = &t
	}
	return
}
