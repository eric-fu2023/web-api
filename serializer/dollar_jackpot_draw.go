package serializer

import (
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/util"
)

type DollarJackpotDraw struct {
	Id                int64          `json:"id"`
	Name              string         `json:"name"`
	Total             *float64       `json:"total,omitempty"`
	Contribution      *float64       `json:"contribution,omitempty"`
	ContributionLimit *float64       `json:"contribution_limit,omitempty"`
	StartTimeTs       int64          `json:"start_time_ts"`
	EndTimeTs         int64          `json:"end_time_ts"`
	IsClosed          bool           `json:"is_closed"`
	DollarJackpot     *DollarJackpot `json:"dollar_jackpot,omitempty"`
	Winner            *SimpleUser    `json:"winner,omitempty"`
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
	if a.Status > 0 {
		b.IsClosed = true
	}
	if a.DollarJackpot != nil {
		t := BuildDollarJackpot(c, *a.DollarJackpot)
		b.DollarJackpot = &t
		prizeFloat := float64(a.DollarJackpot.Prize) / 100
		limit := prizeFloat * model.ContributionLimitPercent
		b.ContributionLimit = &limit
		if b.Total != nil && prizeFloat < *b.Total {
			b.Total = &prizeFloat
		}
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
