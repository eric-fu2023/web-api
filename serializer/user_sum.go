package serializer

import (
	"web-api/model"
)

type UserSum struct {
	Balance        float64 `json:"balance"`
	RemainingWager float64 `json:"wagering_requirement"`
	Withdrawable   float64 `json:"withdrawable"`
}

func BuildUserSum(a model.UserSum) UserSum {
	u := UserSum{
		Balance:        float64(a.Balance) / 100,
		RemainingWager: float64(a.RemainingWager) / 100,
		Withdrawable:   float64(a.MaxWithdrawable) / 100,
	}
	if u.Balance < u.Withdrawable {
		u.Withdrawable = u.Balance
	}
	return u
}
