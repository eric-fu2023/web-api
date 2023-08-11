package model

import (
	"gorm.io/gorm"
)

type GameProviderUser struct {
	Base
	GameProviderId			int64
	UserId					int64
	ExternalUserId			string
	CurrencyGameProviderId	int64
	DeletedAt				gorm.DeletedAt
}

func (GameProviderUser) TableName() string {
	return "game_provider_users"
}