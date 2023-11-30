package serializer

import "web-api/model"

type CashMethod struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	IconURL     string `json:"icon_url"`
	MethodType  string `json:"method_type"`
	BaseURL     string `json:"base_url"`
	CallbackURL string `json:"callback_url"`
	AccountType string `json:"account_type"`
	MinAmount   int64  `json:"min_amount"`
	MaxAmount   int64  `json:"max_amount"`
}

func BuildCashMethodWrapper(minAmount, maxAmount int64) func(a model.CashMethod) CashMethod {
	return func(a model.CashMethod) CashMethod {
		method := BuildCashMethod(a)
		method.MinAmount = minAmount
		method.MaxAmount = maxAmount
		return method
	}
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
	}
}
