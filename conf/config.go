package conf

import (
	"github.com/caarlos0/env/v10"
)

type Config struct {
	WithdrawMinNoDeposit int64   `env:"WITHDRAW_MIN_NO_DEPOSIT"`
	DefaultCurrency      string  `env:"DEFAULT_CURRENCY"`
	ExchangeAdjustment   float64 `env:"EXCHANGE_ADJUSTMENT"`
}

var cfg Config

func InitCfg() {
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
}

func GetCfg() Config {
	return cfg
}
