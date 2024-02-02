package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type Bet struct {
	MatchName  string   `json:"match_name"`
	MatchTime  *int64   `json:"match_time"`
	MarketName string   `json:"market_name"`
	MarketDesc string   `json:"market_desc,omitempty"`
	OptionName string   `json:"option_name"`
	Odds       *float64 `json:"odds,omitempty"`
	Outcome    *int64   `json:"outcome,omitempty"`
}

func BuildBet(c *gin.Context, a ploutos.Bet) (b Bet) {
	b.MarketName = a.GetBetType()
	b.MarketDesc = a.GetBetTypeDesc()
	b.MatchTime = a.GetMatchTime()
	b.OptionName = a.GetBetChoice()
	b.MatchName = a.GetMatchName()
	b.Odds = a.GetOdds()
	b.Outcome = a.GetOutcome()
	return
}
