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
}

func BuildStreamGameSession(c *gin.Context, a ploutos.StreamGameSession) (b StreamGameSession) {
	b = StreamGameSession{
		Id:          a.Id,
		GameId:      a.StreamGameId,
		ReferenceId: a.ReferenceId,
	}
	if a.Result != "" {
		var j map[string]interface{}
		if e := json.Unmarshal([]byte(a.Result), &j); e == nil {
			b.Result = j
		}
	}
	return
}
