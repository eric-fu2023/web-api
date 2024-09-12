package serializer

import (
	"os"

	"blgit.rfdev.tech/taya/payment-service/foray"
	"github.com/shopspring/decimal"
)

func BuildPaymentOrderFromForay(p foray.PaymentOrderRespData, fromCurrency string, fromAmount decimal.Decimal, fromExchangeRate float64, toCurrency string, toAmount decimal.Decimal, toExchangeRate float64) TopupOrder {
	dataType := "H5_URL"

	return TopupOrder{
		TopupOrderNo:     p.OrderNumber,
		TopupOrderStatus: "IN_PROGRESS",
		OrderNumber:      p.OrderTranoIn,
		FromCurrency:     fromCurrency,
		FromAmount:       fromAmount,
		FromExchangeRate: fromExchangeRate,
		ToCurrency:       toCurrency,
		ToAmount:         toAmount,
		ToExchangeRate:   toExchangeRate,
		TopupData:        &p.OrderPayUrl,
		TopupDataType:    &dataType,
		RedirectUrl:      os.Getenv("FORAY_REDIRECT_URL"),
		Html:             *p.OrderPayHtml,
		WalletAddress:    *p.OrderPayInfo,
		BankCardInfo:     paymentBankCardInfo{},
	}
}
