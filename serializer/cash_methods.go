package serializer

import "web-api/model"

type CashMethod struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	IconURL     string `json:"icon_url"`
	MethodType  string `json:"method_type"`
	BaseURL     string `json:"base_url"`
	CallbackURL string `json:"callback_url"`
}

func BuildCashMethod(a model.CashMethod) CashMethod {
	methodType := "top-up"
	if a.MethodType < 0 {
		methodType = "withdraw"
	}
	return CashMethod{
		ID:          a.ID,
		Name:        a.Name,
		IconURL:     a.IconURL,
		MethodType:  methodType,
		BaseURL:     a.BaseURL,
		CallbackURL: a.CallbackURL,
	}
}
