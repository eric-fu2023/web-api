package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type Bet struct {
	MatchName  string `json:"match_name"`
	MatchTime  *int64 `json:"match_time"`
	MarketName string `json:"market_name"`
	OptionName string `json:"option_name"`
	Outcome    *int64 `json:"outcome,omitempty"`
}

func BuildBet(c *gin.Context, a ploutos.Bet) (b Bet) {
	b.MarketName = a.GetBetType()
	b.MatchTime = a.GetMatchTime()
	b.OptionName = a.GetBetChoice()
	b.MatchName = a.GetMatchName()
	b.Outcome = a.GetOutcome()
	return
}
