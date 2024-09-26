package serializer

import (
	"os"

	forayDto "blgit.rfdev.tech/taya/payment-service/foray/dto"
	"github.com/shopspring/decimal"
)

func BuildPaymentOrderFromForay(p forayDto.DepositOrderResponseData, paymentType string, fromCurrency string, fromAmount decimal.Decimal, fromExchangeRate float64, toCurrency string, toAmount decimal.Decimal, toExchangeRate float64) TopupOrder {
	dataType := retrievePaymentDataType(paymentType)

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
		Html:             p.GetHtml(),
		WalletAddress:    p.GetWallet(),
		BankCardInfo:     paymentBankCardInfo{},
	}
}

func retrievePaymentDataType(paymentType string) string {
	switch paymentType {
	case "CRYPTO":
		return "CRYPTO_WALLET"
	default:
		return "H5_URL"
	}
}
