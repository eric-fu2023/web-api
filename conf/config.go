package conf

import (
	"github.com/caarlos0/env/v10"
)

type Config struct {
	FirstTopupMinimum int64 `env:"FIRST_TOPUP_MINIMUM"`
	TopupMinimum      int64 `env:"TOPUP_MINIMUM"`
	TopupMax          int64 `env:"TOPUP_MAX"`
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
