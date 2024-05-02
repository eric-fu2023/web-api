package exchange

import (
	"context"
	"fmt"
	"strconv"
	"web-api/cache"
	"web-api/conf"
)

const (
	RedisExchangeRateKey = "OKX_EXCHANGE_RATE:%s-%s"
)

type OkxClient struct{}

func (k OkxClient) GetExchangeRate(c context.Context, sourceCurrency, destCurrency string) (ExchangeRates, error) {
	if sourceCurrency == destCurrency || (sourceCurrency == USD && destCurrency == USDT) || (sourceCurrency == USDT && destCurrency == USD) {
		return ExchangeRates{
			ExchangeRate:         1,
			AdjustedExchangeRate: 1,
		}, nil
	}
	res := cache.RedisClient.Get(c, fmt.Sprintf(RedisExchangeRateKey, sourceCurrency, destCurrency))
	if res.Err() != nil {
		// var resp OkxResponse
		// resty.New().SetDebug(true).SetBaseURL("https://www.okx.com/").R().
		// 	SetResult(&resp).
		// 	Get("/api/v5/market/exchange-rate")
		// val = resp.Data[0][usdCny]
		return ExchangeRates{}, res.Err()
	}
	val := res.Val()
	rate, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return ExchangeRates{}, err
	}
	ret := ExchangeRates{
		ExchangeRate:         rate,
		AdjustedExchangeRate: rate * (1 - conf.GetCfg().ExchangeAdjustment),
	}
	return ret, nil
}

// type OkxResponse struct {
// 	Code string              `json:"code"`
// 	Msg  string              `json:"msg"`
// 	Data []map[string]string `json:"data"`
// }

// const usdCny = "usdCny"
