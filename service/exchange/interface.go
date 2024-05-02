package exchange

import "context"



type ExchangeInterface interface {
	GetExchangeRate(c context.Context, sourceCurrency, destCurrency string) (ExchangeRates, error)
}
