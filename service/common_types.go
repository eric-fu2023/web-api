package service

import "web-api/model"

type Platform struct {
	Platform int64 `form:"platform" json:"platform" binding:"required"`
}

type Page struct {
	Page  int `form:"page" json:"page" binding:"min=1"`
	Limit int `form:"limit" json:"limit" binding:"min=1"`
}

func GetGameProviderUser(provider int64, userId string) (gpu model.GameProviderUser, err error) {
	err = gpu.GetByProviderAndExternalUser(provider, userId)
	return
}

func GetSums(gpu model.GameProviderUser) (balance int64, remainingWager int64, maxWithdrawable int64, err error) {
	var userSum model.UserSum
	err = model.DB.Where(`user_id`, gpu.UserId).First(&userSum).Error
	if err != nil {
		return
	}
	balance = userSum.Balance
	remainingWager = userSum.RemainingWager
	maxWithdrawable = userSum.MaxWithdrawable
	return
}
