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
	StartTimeTs   int64          `json:"start_time_ts"`
	EndTimeTs     int64          `json:"end_time_ts"`
	DollarJackpot *DollarJackpot `json:"dollar_jackpot,omitempty"`
	Winner        *SimpleUser    `json:"winner,omitempty"`
}

func BuildDollarJackpotDraw(c *gin.Context, a model.DollarJackpotDraw) (b DollarJackpotDraw) {
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
	return
}
