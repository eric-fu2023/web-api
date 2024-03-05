package serializer

import (
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/util"
)

type DollarJackpotDraw struct {
	Id            int64          `json:"id"`
	Name          string         `json:"name"`
	Total         *float64       `json:"total,omitempty"`
	Contribution  *float64       `json:"contribution,omitempty"`
	StartTimeTs   int64          `json:"start_time_ts"`
	EndTimeTs     int64          `json:"end_time_ts"`
	DollarJackpot *DollarJackpot `json:"dollar_jackpot,omitempty"`
	Winner        *SimpleUser    `json:"winner,omitempty"`
}

func BuildDollarJackpotDraw(c *gin.Context, a model.DollarJackpotDraw, contribution *int64) (b DollarJackpotDraw) {
	b = DollarJackpotDraw{
		Id:          a.ID,
		Name:        a.Name,
		StartTimeTs: a.StartTime.Unix(),
		EndTimeTs:   a.EndTime.Unix(),
	}
	if a.Total != nil {
		t := util.MoneyFloat(*a.Total)
		b.Total = &t
	}
	if a.DollarJackpot != nil {
		t := BuildDollarJackpot(c, *a.DollarJackpot)
		b.DollarJackpot = &t
	}
	if a.Winner.ID != 0 {
		t := BuildSimpleUser(c, a.Winner)
		b.Winner = &t
	}
	if contribution != nil {
		t := util.MoneyFloat(*contribution)
		b.Contribution = &t
	}
	return
}
