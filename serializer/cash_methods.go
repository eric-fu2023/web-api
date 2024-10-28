package serializer

import (
	"web-api/model"
	"web-api/util"
)

type DefaultCashMethodPromotionSelection struct {
	SelectionAmount     float64 `json:"selection_amount"`
	Label               string  `json:"label"`
	Icon                string  `json:"icon"`
	BonusRate           float64 `json:"bonus_rate"`
	BonusAmount         float64 `json:"bonus_amount"`
	NeedCustomerSupport bool    `json:"need_customer_support"`
}

type CashMethodPromotion struct {
	PayoutRate         float64 `json:"payout_rate"`
	MaxPromotionAmount float64 `json:"max_promotion_amount"`
	MinAmountForPayout float64 `json:"min_payout"`

	DefaultCashMethodPromotionSelections []DefaultCashMethodPromotionSelection `json:"cash_method_promotion_selections"`
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

// BuildCashMethod
// Deprecated. refer and use [BuildCashMethodWithPromotion]/[BuildCashMethod2]
func BuildCashMethod(a model.CashMethod, maxClaimableByCashMethodId MaxPromotionAmountByCashMethodId) CashMethod {
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
			PayoutRate:                           a.CashMethodPromotion.PayoutRate,
			MaxPromotionAmount:                   float64(maxClaimableByCashMethodId[a.ID]) / 100,
			DefaultCashMethodPromotionSelections: nil,
			MinAmountForPayout:                   float64(a.CashMethodPromotion.MinPayout) / 100,
		}
	}

	return cashMethod
}

func BuildCashMethod2(cm model.CashMethod) CashMethod {
	methodType := "top-up"
	if cm.MethodType < 0 {
		methodType = "withdraw"
	}

	cashMethod := CashMethod{
		ID:          cm.ID,
		Name:        cm.Name,
		IconURL:     Url(cm.IconURL),
		MethodType:  methodType,
		BaseURL:     cm.BaseURL,
		CallbackURL: cm.CallbackURL,
		AccountType: cm.AccountType,
		MinAmount:   cm.MinAmount / 100,
		MaxAmount:   cm.MaxAmount / 100,
		DefaultOptions: util.MapSlice(cm.DefaultOptions, func(option int32) int {
			return int(option) / 100
		}),
		Currency:            cm.Currency,
		AccountNameRequired: cm.AccountType == "bank_card",
	}

	return cashMethod
}

func BuildCashMethodWithCashMethodPromotion(_cm model.CashMethod, cashMethodPromotion *CashMethodPromotion) CashMethod {
	methodType := "top-up"
	if _cm.MethodType < 0 {
		methodType = "withdraw"
	}

	cashMethod := CashMethod{
		ID:          _cm.ID,
		Name:        _cm.Name,
		IconURL:     Url(_cm.IconURL),
		MethodType:  methodType,
		BaseURL:     _cm.BaseURL,
		CallbackURL: _cm.CallbackURL,
		AccountType: _cm.AccountType,
		MinAmount:   _cm.MinAmount / 100,
		MaxAmount:   _cm.MaxAmount / 100,
		DefaultOptions: util.MapSlice(_cm.DefaultOptions, func(option int32) int {
			return int(option) / 100
		}),
		Currency:            _cm.Currency,
		AccountNameRequired: _cm.AccountType == "bank_card",
	}

	cashMethod.CashMethodPromotion = cashMethodPromotion
	return cashMethod
}
