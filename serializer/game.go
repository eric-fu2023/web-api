package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type Game struct {
	GameName   string `json:"game_name"`
	SeasonName string `json:"season_name,omitempty"`
}

func BuildGame(c *gin.Context, a ploutos.Game) (b Game) {
	b = Game{
		GameName:   a.GetGameName(),
		SeasonName: a.GetSeasonName(),
	}
	return
}
