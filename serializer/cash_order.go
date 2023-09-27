package serializer

import (
	"time"
	"web-api/conf/consts"
	"web-api/model"

	"blgit.rfdev.tech/taya/payment-service/finpay"
)

type TopupOrder struct {
	TopupOrderNo     string  `json:"topup_order_no"`
	TopupOrderStatus string  `json:"topup_order_status"`
	OrderNumber      string  `json:"order_number"`
	TopupData        *string `json:"topup_data"`
	TopupDataType    *string `json:"topup_data_type"`
}

func BuildPaymentOrder(p finpay.PaymentOrderRespData) TopupOrder {
	return TopupOrder{
		TopupOrderNo:     p.PaymentOrderNo,
		TopupOrderStatus: p.PaymentOrderStatus,
		OrderNumber:      p.MerchantOrderNo,
		TopupData:        p.PaymentData,
		TopupDataType:    p.PaymentDataType,
	}
}

type WithdrawOrder struct {
	WithdrawOrderNo     string `json:"withdraw_order_no"`
	WithdrawOrderStatus string `json:"withdraw_order_status"`
	OrderNumber         string `json:"order_number"`
}

func BuildWithdrawOrder(p model.CashOrder) WithdrawOrder {
	return WithdrawOrder{
		WithdrawOrderNo:     p.TransactionId,
		WithdrawOrderStatus: consts.CashOrderStatus[p.Status],
		OrderNumber:         p.ID,
	}
}

type GenericCashOrder struct {
	OrderNo         string    `json:"order_no"`
	OrderStatus     string    `json:"order_status"`
	CreatedAt       time.Time `json:"created_at"`
	Amount          int64     `json:"amount"`
	EffectiveAmount int64     `json:"effective_amount"`
	OrderType       string    `json:"order_type"`
	// Currency    string    `json:"currency"`
}

func BuildGenericCashOrder(p model.CashOrder) GenericCashOrder {
	amount := p.AppliedCashInAmount
	effectiveAmount := p.EffectiveCashInAmount
	if p.OrderType < 0 {
		amount = p.AppliedCashOutAmount
		effectiveAmount = p.EffectiveCashOutAmount
	}

	return GenericCashOrder{
		OrderNo:         p.ID,
		OrderStatus:     consts.CashOrderStatus[p.Status],
		CreatedAt:       p.CreatedAt,
		Amount:          amount,
		EffectiveAmount: effectiveAmount,
	}
}
