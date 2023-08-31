package model

import models "blgit.rfdev.tech/taya/ploutos-object"

type Currency models.CurrencyC
type CurrencyGameProvider models.CurrencyGameProviderC
type GameProviderUser models.GameProviderUserC
type Transaction models.TransactionC
type FbTransaction models.FbTransactionC
type UserSum models.UserSumC

func(a *GameProviderUser) GetByProviderAndExternalUser(provider int64, userId string) error {
	return DB.Where(`game_provider_id`, provider).Where(`external_user_id`, userId).First(&a).Error
}