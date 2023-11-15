package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type Game struct {
	GameName string `json:"game_name"`
}

func BuildGame(c *gin.Context, a ploutos.Game) (b Game) {
	b.GameName = a.GetGameName()
	return
}
