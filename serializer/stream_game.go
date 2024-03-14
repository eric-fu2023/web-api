package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"web-api/util"
)

type StreamGame struct {
	Id         int64   `json:"id"`
	Name       string  `json:"game_name"`
	MinBet     float64 `json:"min_bet"`
	MaxBet     float64 `json:"max_bet"`
	RuleImgWeb string  `json:"rule_img_web"`
	RuleImgH5  string  `json:"rule_img_h5"`
}

func BuildStreamGame(c *gin.Context, a ploutos.StreamGame) (b StreamGame) {
	b = StreamGame{
		Id:         a.ID,
		Name:       a.Name,
		MinBet:     util.MoneyFloat(a.MinBet),
		MaxBet:     util.MoneyFloat(a.MaxBet),
		RuleImgWeb: Url(a.RuleImgWeb),
		RuleImgH5:  Url(a.RuleImgH5),
	}
	return
}
