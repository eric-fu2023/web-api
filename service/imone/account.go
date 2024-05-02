package imone

import (
	"errors"

	"web-api/model"
	"web-api/util"

	imonegameservice "blgit.rfdev.tech/taya/game-service/imone"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

func (c *ImOne) CreateUser(user model.User, currency string) error {
	err := c.createImOneUserAndDbWallet(user, currency)

	if errors.As(err, &imonegameservice.ErrCreateUserAlreadyExists{}) || errors.Is(err, gorm.ErrDuplicatedKey) {
		err = c.VendorRegisterError()
	}

	return err
}

func (c *ImOne) CreateWallet(user model.User, currency string) error {
	return c.createImOneUserAndDbWallet(user, currency)
}

func (c *ImOne) createImOneUserAndDbWallet(user model.User, currency string) error {
	err := util.ImOneFactory().CreateUser(user.IdAsString(), currency, user.Password, "")
	if err != nil {
		return err
	}

	return model.DB.Transaction(func(tx *gorm.DB) (err error) {
		var gameVendors []ploutos.GameVendor
		err = tx.Model(ploutos.GameVendor{}).
			Where(`game_vendor.game_integration_id`, util.IntegrationIdImOne).Find(&gameVendors).Error
		if err != nil {
			return
		}

		for _, gameVendor := range gameVendors {
			gvu := ploutos.GameVendorUser{
				GameVendorId:     gameVendor.ID,
				UserId:           user.ID,
				ExternalUserId:   user.Username,
				ExternalCurrency: currency,
			}

			err = tx.Create(&gvu).Error
			if err != nil {
				return
			}
		}
		return
	})
}

func (c *ImOne) GetGameBalance(user model.User, currency, gameCode string, _ model.Extra) (balance int64, _err error) {
	productWalletCode, exist := tayaGameCodeToImOneWalletCodeMapping[gameCode]

	if !exist {
		return 0, errors.New("unknown game code")
	}

	client := util.ImOneFactory()
	balanceFloat, err := client.GetWalletBalance(user.Username, productWalletCode)
	if err != nil {
		return 0, errors.Join(errors.New("ImOne get balance error"), err)
	}

	return util.MoneyInt(balanceFloat), nil
}
