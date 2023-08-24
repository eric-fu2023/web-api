package model

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
)

type GameProviderUser struct {
	models.GameProviderUserC
}

func(a *GameProviderUser) GetByProviderAndExternalUser(provider int64, userId string) error {
	return DB.Where(`game_provider_id`, provider).Where(`external_user_id`, userId).First(&a).Error
}