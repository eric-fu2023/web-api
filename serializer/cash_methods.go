package serializer

import (
	"web-api/model"
	"web-api/util"
)

type CashMethodPromotion struct {
	PayoutRate                    float64   `json:"payout_rate"`
	MaxPromotionAmount            float64   `json:"max_promotion_amount"`
	DefaultOptionPromotionAmounts []float64 `json:"default_option_promotion_amounts"`
	MinAmountForPayout            float64   `json:"min_payout"`
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

type PromotionAmountByCashMethodId = map[int64]int64
type MaxPromotionAmountByCashMethodId = PromotionAmountByCashMethodId

func BuildCashMethod(a model.CashMethod, maxPromotionAmountByCashMethodId MaxPromotionAmountByCashMethodId) CashMethod {
	methodType := "top-up"
	if a.MethodType < 0 {
		methodType = "withdraw"
	}
	cashMethod := CashMethod{
		ID:          a.ID,
		Name:        a.Name,
		IconURL:     Url(a.IconURL),
		MethodType:  methodType,
		BaseURL:     a.BaseURL,
		CallbackURL: a.CallbackURL,
		AccountType: a.AccountType,
		MinAmount:   a.MinAmount / 100,
		MaxAmount:   a.MaxAmount / 100,
		DefaultOptions: util.MapSlice(a.DefaultOptions, func(option int32) int {
			return int(option) / 100
		}),
		Currency:            a.Currency,
		AccountNameRequired: a.AccountType == "bank_card",
	}

	if a.CashMethodPromotion != nil {
		cashMethod.CashMethodPromotion = &CashMethodPromotion{
			PayoutRate:         a.CashMethodPromotion.PayoutRate,
			MaxPromotionAmount: float64(maxPromotionAmountByCashMethodId[a.ID]) / 100,
			DefaultOptionPromotionAmounts: util.MapSlice(a.DefaultOptions, func(defaultOption int32) (amount float64) {
				amount = float64(defaultOption) * a.CashMethodPromotion.PayoutRate
				maxAmount, exist := maxPromotionAmountByCashMethodId[a.ID]
				if exist && amount > float64(maxAmount) {
					amount = float64(maxAmount)
				}
				return amount / 100
			}),
			MinAmountForPayout: float64(a.CashMethodPromotion.MinPayout) / 100,
		}
	}

	return cashMethod
}
