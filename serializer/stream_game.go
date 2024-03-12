package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type StreamGame struct {
	Id   int64  `json:"id"`
	Name string `json:"game_name"`
}

func BuildStreamGame(c *gin.Context, a ploutos.StreamGame) (b StreamGame) {
	b = StreamGame{
		Id:   a.ID,
		Name: a.Name,
	}
	return
}
