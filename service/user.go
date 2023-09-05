package service

import (
	"errors"
	"fmt"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/util"
)

var (
	ErrEmptyCurrencyId    = errors.New("empty currency id")
	ErrFbCreateUserFailed = errors.New("fb create user failed")
)

func CreateUser(user model.User) error {
	tx := model.DB.Begin()
	err := tx.Save(&user).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	userSum := model.UserSum{
		UserId: user.ID,
	}
	err = tx.Create(&userSum).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	var currency model.CurrencyGameProvider
	err = model.DB.Where(`game_provider_id`, consts.GameProvider["fb"]).Where(`currency_id`, user.CurrencyId).First(&currency).Error
	if err != nil {
		tx.Rollback()
		return ErrEmptyCurrencyId
	}
	client := util.FBFactory.NewClient()
	res, err := client.CreateUser(user.Username, []int64{}, 0)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%w: %w", ErrFbCreateUserFailed, err)
	}
	gpu := model.GameProviderUser{
		GameProviderId:     consts.GameProvider["fb"],
		UserId:             user.ID,
		ExternalUserId:     user.Username,
		ExternalCurrencyId: currency.Value,
		ExternalId:         fmt.Sprintf("%d", res),
	}
	err = tx.Save(&gpu).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%w: %w", ErrFbCreateUserFailed, err)
	}
	tx = tx.Commit()

	return nil
}
