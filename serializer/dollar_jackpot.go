package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"web-api/util"
)

type DollarJackpot struct {
	Name  string  `json:"name"`
	Prize float64 `json:"prize"`
}

func BuildDollarJackpot(a ploutos.DollarJackpot) (b DollarJackpot) {
	b = DollarJackpot{
		Name:  a.Name,
		Prize: util.MoneyFloat(a.Prize),
	}
	return
}
