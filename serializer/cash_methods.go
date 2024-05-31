package serializer

import (
	"web-api/model"
	"web-api/util"
)

type CashMethod struct {
	ID                  int64  `json:"id"`
	Name                string `json:"name"`
	IconURL             string `json:"icon_url"`
	MethodType          string `json:"method_type"`
	BaseURL             string `json:"base_url"`
	CallbackURL         string `json:"callback_url"`
	AccountType         string `json:"account_type"`
	MinAmount           int64  `json:"min_amount"`
	MaxAmount           int64  `json:"max_amount"`
	DefaultOptions      []int  `json:"default_options"`
	Currency            string `json:"currency"`
	AccountNameRequired bool   `json:"account_name_required"`
}

func BuildCashMethod(a model.CashMethod) CashMethod {
	methodType := "top-up"
	if a.MethodType < 0 {
		methodType = "withdraw"
	}
	return CashMethod{
		ID:          a.ID,
		Name:        a.Name,
		IconURL:     Url(a.IconURL),
		MethodType:  methodType,
		BaseURL:     a.BaseURL,
		CallbackURL: a.CallbackURL,
		AccountType: a.AccountType,
		MinAmount:   a.MinAmount / 100,
		MaxAmount:   a.MaxAmount / 100,
		DefaultOptions: util.MapSlice(a.DefaultOptions, func(a int32) int {
			return int(a) / 100
		}),
		Currency:            a.Currency,
		AccountNameRequired: a.AccountType == "bank_card",
	}
}
