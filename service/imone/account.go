package imone

import (
	"errors"

	"web-api/model"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

func (c *ImOne) CreateWallet(user model.User, currency string) error {
	return c.createImOneUserAndDbWallet(user, currency)
}

const defaultPassword = "qq123456"

// skips duplicate error on insert.
func (c *ImOne) createImOneUserAndDbWallet(user model.User, currency string) error {
	// FIXME password to be derived from user instead of default value
	go func() {
		// fire and forget. later calls should follow up with user creation, if needed.
		_ = util.ImOneFactory().CreateUser(user.IdAsString(), currency, defaultPassword, "")
	}()

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
				ExternalUserId:   user.IdAsString(),
				ExternalCurrency: currency,
			}

			err = tx.Create(&gvu).Error
			if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
				return
			}
		}
		return
	})
}

func (c *ImOne) GetGameBalance(user model.User, currency, gameCode string, extra model.Extra) (balance int64, _err error) {
	productWalletCode, exist := tayaGameCodeToImOneWalletCodeMapping[gameCode]
	if !exist {
		return 0, ErrGameCodeMapping
	}

	client := util.ImOneFactory()
	balanceFloat, err := client.GetWalletBalance(user.IdAsString(), productWalletCode)
	if err != nil {
		return 0, errors.Join(ErrGetBalance, err)
	}

	return util.MoneyInt(balanceFloat), nil
}
