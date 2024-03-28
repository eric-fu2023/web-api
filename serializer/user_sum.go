package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type UserSum struct {
	Balance        float64 `json:"balance"`
	RemainingWager float64 `json:"wagering_requirement"`
	Withdrawable   float64 `json:"withdrawable"`
}

func BuildUserSum(a ploutos.UserSum) UserSum {
	var withdrawable float64
	if a.RemainingWager == 0 {
		withdrawable = float64(a.Balance) / 100
	} else {
		withdrawable = 0
	}
	u := UserSum{
		Balance:        float64(a.Balance) / 100,
		RemainingWager: float64(a.RemainingWager) / 100,
		Withdrawable:   withdrawable,
	}
	return u
}
