package exchange

import "context"

const (
	RedisExchangeRateKey = "OKX_EXCHANGE_RATE:%s-%s"
)

type ExchangeInterface interface {
	GetExchangeRate(c context.Context, sourceCurrency, destCurrency string) (ExchangeRates, error)
}
