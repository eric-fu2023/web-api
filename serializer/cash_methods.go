package serializer

import (
	"web-api/model"
	"web-api/util"
)

type CashMethodPromotion struct {
	PayoutRate                    float64   `json:"payout_rate"`
	MaxPromotionAmount            int64     `json:"max_promotion_amount"`
	DefaultOptionPromotionAmounts []float64 `json:"default_option_promotion_amounts"`
}

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

	CashMethodPromotion *CashMethodPromotion `json:"cash_method_promotion,omitempty"`
}

func BuildCashMethod(a model.CashMethod, maxPromotionAmountByCashMethodId map[int64]int64) (cashMethod CashMethod) {
	methodType := "top-up"
	if a.MethodType < 0 {
		methodType = "withdraw"
	}
	cashMethod = CashMethod{
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

	if a.CashMethodPromotion != nil {
		cashMethod.CashMethodPromotion = &CashMethodPromotion{
			PayoutRate:         a.CashMethodPromotion.PayoutRate,
			MaxPromotionAmount: maxPromotionAmountByCashMethodId[a.ID] / 100,
			DefaultOptionPromotionAmounts: util.MapSlice(a.DefaultOptions, func(b int32) (amount float64) {
				amount = float64(b) * a.CashMethodPromotion.PayoutRate
				maxAmount, exist := maxPromotionAmountByCashMethodId[a.ID]
				if exist && amount > float64(maxAmount) {
					amount = float64(maxAmount)
				}
				return amount / 100
			}),
		}
	}

	return cashMethod
}
