package serializer

import (
	"os"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	"github.com/shopspring/decimal"
)

func buildPaymentBankCardInfo(p finpay.PaymentBankCardInfo) paymentBankCardInfo {
	return paymentBankCardInfo{
		Amount:          float64(p.Amount) / 100,
		BankCode:        p.BankCode,
		BankName:        p.BankName,
		BankAccountNo:   p.BankAccountNo,
		BankBranchName:  p.BankBranchName,
		BankAccountName: p.BankAccountName,
	}
}

func BuildPaymentOrderFromFinpay(p finpay.PaymentOrderRespData, fromCurrency string, fromAmount decimal.Decimal, fromExchangeRate float64, toCurrency string, toAmount decimal.Decimal, toExchangeRate float64) TopupOrder {
	d := p.GetUrl()
	b := buildPaymentBankCardInfo(p.GetBankInfo())

	return TopupOrder{
		TopupOrderNo:     p.PaymentOrderNo,
		TopupOrderStatus: p.PaymentOrderStatus,
		OrderNumber:      p.MerchantOrderNo,
		FromCurrency:     fromCurrency,
		FromAmount:       fromAmount,
		FromExchangeRate: fromExchangeRate,
		ToCurrency:       toCurrency,
		ToAmount:         toAmount,
		ToExchangeRate:   toExchangeRate,
		TopupData:        &d,
		TopupDataType:    &p.PaymentDataType,
		RedirectUrl:      os.Getenv("FINPAY_REDIRECT_URL"),
		Html:             p.GetHtml(),
		WalletAddress:    p.GetWallet(),
		BankCardInfo:     b,
	}
}
