package serializer

import (
	"web-api/model"
	"web-api/util"

	"github.com/gin-gonic/gin"
)

type DollarJackpotBetReportResponse struct {
	Id          *string  `json:"id"`
	Name        string   `json:"name"`
	Total       *float64 `json:"total"`
	StartTimeTs int64    `json:"start_time_ts"`
	EndTimeTs   int64    `json:"end_time_ts"`
	Status      int64    `json:"status"` // 0:ongoing, 1: computing, 2: ended
	IsWin       bool     `json:"is_win"`
	Win         float64    `json:"win"`
}

func BuildDollarJackpotBetReportResponse(c *gin.Context, a model.DollarJackpotBetReport, contribution *int64) (b DollarJackpotBetReportResponse) {
	b = DollarJackpotBetReportResponse{
		Id:          a.ID,
		Name:        a.JackpotDraws.Name,
		StartTimeTs: a.JackpotDraws.StartTime.Unix(),
		EndTimeTs:   a.JackpotDraws.EndTime.Unix(),
		Status:      a.JackpotDraws.Status,
		IsWin:       a.Win > 0,
		Win:         float64(a.Win - a.Bet)/100.0,
	}
	if a.JackpotDraws.DollarJackpot != nil {
		t := util.MoneyFloat(a.JackpotDraws.DollarJackpot.Prize)
		b.Total = &t
	}
	return
}
