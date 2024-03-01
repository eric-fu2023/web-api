package serializer

import (
	"web-api/model"
	"web-api/util"
)

type DollarJackpotDraw struct {
	Id            int64          `json:"id"`
	Total         float64        `json:"total"`
	End           int64          `json:"end_ts"`
	DollarJackpot *DollarJackpot `json:"dollar_jackpot,omitempty"`
}

func BuildDollarJackpotDraw(a model.DollarJackpotDraw) (b DollarJackpotDraw) {
	b = DollarJackpotDraw{
		Id:    a.ID,
		Total: util.MoneyFloat(a.Total),
		End:   a.EndTime.Unix(),
	}
	if a.DollarJackpot != nil {
		t := BuildDollarJackpot(*a.DollarJackpot)
		b.DollarJackpot = &t
	}
	return
}
