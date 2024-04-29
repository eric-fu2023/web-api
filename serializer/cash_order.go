package serializer

import (
	"os"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/util/i18n"

	"blgit.rfdev.tech/taya/payment-service/finpay"
)

type TopupOrder struct {
	TopupOrderNo     string  `json:"topup_order_no"`
	TopupOrderStatus string  `json:"topup_order_status"`
	OrderNumber      string  `json:"order_number"`
	TopupData        *string `json:"topup_data"`
	TopupDataType    *string `json:"topup_data_type"`
	RedirectUrl      string  `json:"redirect_url"`
	Html             string  `json:"html"`
	WalletAddress    string  `json:"wallet_address"`
}

func BuildPaymentOrder(p finpay.PaymentOrderRespData) TopupOrder {
	d := p.GetUrl()
	return TopupOrder{
		TopupOrderNo:     p.PaymentOrderNo,
		TopupOrderStatus: p.PaymentOrderStatus,
		OrderNumber:      p.MerchantOrderNo,
		TopupData:        &d,
		TopupDataType:    &p.PaymentDataType,
		RedirectUrl:      os.Getenv("FINPAY_REDIRECT_URL"),
		Html:             p.GetHtml(),
		WalletAddress:    p.GetWallet(),
	}
}

type WithdrawOrder struct {
	WithdrawOrderNo     string `json:"withdraw_order_no"`
	WithdrawOrderStatus string `json:"withdraw_order_status"`
	OrderNumber         string `json:"order_number"`
}

func BuildWithdrawOrder(p model.CashOrder) WithdrawOrder {
	txnID := ""
	if p.TransactionId != nil {
		txnID = *p.TransactionId
	}
	return WithdrawOrder{
		WithdrawOrderNo:     txnID,
		WithdrawOrderStatus: consts.CashOrderStatus[p.Status],
		OrderNumber:         p.ID,
	}
}

type GenericCashOrder struct {
	OrderNo         string  `json:"order_no"`
	OrderStatus     string  `json:"order_status"`
	CreatedAt       int64   `json:"created_at"`
	Amount          float64 `json:"amount"`
	EffectiveAmount float64 `json:"effective_amount"`
	OrderType       string  `json:"order_type"`
	// Currency    string    `json:"currency"`
	TypeDetail string  `json:"type_detail"`
	Wager      float64 `json:"wager"`
}

func BuildGenericCashOrder(p model.CashOrder, i18n i18n.I18n) GenericCashOrder {
	amount := p.AppliedCashInAmount
	effectiveAmount := p.EffectiveCashInAmount
	orderType := consts.OrderTypeMap[p.OrderType]
	detail := i18n.T(consts.OrderOperationTypeDetailMap[p.OperationType])
	if p.OperationType == 0 {
		detail = i18n.T(consts.OrderTypeDetailMap[p.OrderType])
	}
	if p.OrderType < 0 {
		amount = p.AppliedCashOutAmount
		effectiveAmount = p.EffectiveCashOutAmount
	}

	return GenericCashOrder{
		OrderNo:         p.ID,
		OrderStatus:     consts.CashOrderStatus[p.Status],
		CreatedAt:       p.CreatedAt.Unix(),
		Amount:          float64(amount) / 100,
		EffectiveAmount: float64(effectiveAmount) / 100,
		OrderType:       orderType,
		TypeDetail:      detail,
		Wager:           float64(p.WagerChange) / 100,
	}
}
