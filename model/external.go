package model

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

type Currency models.CurrencyC
type CurrencyGameProvider models.CurrencyGameProviderC
type GameProviderUser models.GameProviderUserC
type Transaction models.TransactionC
type FbTransaction models.FbTransactionC
type SabaTransaction models.SabaTransactionC
type UserSum models.UserSumC
type AppConfig models.AppConfigC

func (a *GameProviderUser) GetByProviderAndExternalUser(provider int64, userId string) error {
	return DB.Where(`game_provider_id`, provider).Where(`external_user_id`, userId).First(&a).Error
}

func ByBrandAgentDeviceAndKey(brand int64, agent int64, device int64, key string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		q := db.Where(`brand_id = ? OR brand_id = 0`, brand).Where(`agent_id = ? OR agent_id = 0`, agent).Where(`platform = ? OR platform = 0`, device)
		if key != "" {
			q = q.Where(`key`, key)
		}
		return q
	}
}
