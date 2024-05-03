package imone

import (
	"errors"
	"log"

	"web-api/model"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

func (c *ImOne) CreateWallet(user model.User, currency string) error {
	return c.createImOneUserAndDbWallet(user, currency)
}

const defaultPassword = "qq123456"

func (c *ImOne) createImOneUserAndDbWallet(user model.User, currency string) error {
	// FIXME password to be derived from user instead of default value
	err := util.ImOneFactory().CreateUser(user.IdAsString(), currency, defaultPassword, "")
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

func (c *ImOne) GetGameBalance(user model.User, currency, gameCode string, extra model.Extra) (balance int64, _err error) {
	log.Printf("(c *ImOne) GetGameBalance user %+v currency %+v gameCode %+v extra %+v\n", user, currency, gameCode, extra)
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
