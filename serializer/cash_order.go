package serializer

import "blgit.rfdev.tech/taya/payment-service/finpay"

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
