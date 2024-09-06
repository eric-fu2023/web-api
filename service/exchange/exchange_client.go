package exchange

import (
	"context"
	"fmt"
	"strconv"
	"web-api/cache"
	"web-api/conf"
	"web-api/util"
)

const (
	RedisExchangeRateKey = "OKX_EXCHANGE_RATE:%s-%s"
)

type ExchangeClient struct {
}

type ExchangeRates struct {
	ExchangeRate         float64
	AdjustedExchangeRate float64
}

func (ex ExchangeClient) GetExchangeRate(c context.Context, destCurrency string, isCashIn bool) (ExchangeRates, error) {
	fromCurrency := conf.GetCfg().DefaultCurrency
	toCurrency := destCurrency
	if toCurrency == USDT {
		toCurrency = USD
	}
	if toCurrency == fromCurrency {
		return ExchangeRates{
			ExchangeRate:         1,
			AdjustedExchangeRate: 1,
		}, nil
	}
	res := cache.RedisClient.Get(c, fmt.Sprintf(RedisExchangeRateKey, fromCurrency, toCurrency))
	if res.Err() != nil {
		util.Log().Error("Failed to retrieve exchange rate from %s to %s:", fromCurrency, toCurrency, res.Err())
		return ExchangeRates{}, res.Err()
	}
	val := res.Val()
	rate, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return ExchangeRates{}, err
	}
	ret := ExchangeRates{
		ExchangeRate: rate,
	}
	if isCashIn {
		ret.AdjustedExchangeRate = rate * (1 + conf.GetCfg().ExchangeAdjustment)
	} else {
		ret.AdjustedExchangeRate = rate * (1 - conf.GetCfg().ExchangeAdjustment)
	}
	return ret, nil
}
