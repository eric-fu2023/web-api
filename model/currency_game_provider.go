package model

import (
	"gorm.io/gorm"
)

type CurrencyGameProvider struct {
	Base
	CurrencyId		int64
	GameProviderId	int64
	Value			int64
	DeletedAt		gorm.DeletedAt
}

func (CurrencyGameProvider) TableName() string {
	return "currency_game_providers"
}