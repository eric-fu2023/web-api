package mumbai

import (
	"fmt"
	"web-api/model"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

func (c *Mumbai) CreateWallet(user model.User, currency string) error {
	// create a record for the user for mumbai game.
	return model.DB.Transaction(func(tx *gorm.DB) (err error) {
		var gameVendors []ploutos.GameVendor
		err = tx.Model(ploutos.GameVendor{}).Joins(`INNER JOIN game_vendor_brand gvb ON gvb.game_vendor_id = game_vendor.id`).
			Where(`game_vendor.game_integration_id`, util.IntegrationIdMumbai).Find(&gameVendors).Error

		if err != nil {
			return
		}

		externalUserId := c.Merchant + c.Agent + user.IdAsString()

		for _, gameVendor := range gameVendors {
			gvu := ploutos.GameVendorUser{
				GameVendorId:     gameVendor.ID,
				UserId:           user.ID,
				ExternalUserId:   externalUserId, // USE THE USERNAME WITH SPECIFIC RULE
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

func (c *Mumbai) GetGameBalance(user model.User, currency, gameCode string, extra model.Extra) (balance int64, _err error) {
	// create the client to call the web-service.
	client, err := util.MumbaiFactory()
	if err != nil {
		return 0, err
	}

	username := c.Merchant + c.Agent + fmt.Sprintf("%08s", user.IdAsString())
	balanceFloat, err := client.CheckBalanceUser(username)
	if err != nil {
		return 0, ErrGetBalance
	}
	return util.MoneyInt(balanceFloat), nil
}
