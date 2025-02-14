package mumbai

import (
	"context"
	"fmt"
	"os"
	"web-api/model"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

func (c *Mumbai) CreateWallet(ctx context.Context, user model.User, currency string) error {
	// create a record for the user for mumbai game.
	return model.DB.Transaction(func(tx *gorm.DB) (err error) {
		var gameVendors []ploutos.GameVendor
		err = tx.Model(ploutos.GameVendor{}).Joins(`INNER JOIN game_vendor_brand gvb ON gvb.game_vendor_id = game_vendor.id`).
			Where(`game_vendor.game_integration_id`, util.IntegrationIdMumbai).Find(&gameVendors).Error

		if err != nil {
			return
		}

		externalUserId := os.Getenv("GAME_MUMBAI_MERCHANT_CODE") + os.Getenv("GAME_MUMBAI_AGENT_CODE") + user.IdAsString()

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

func (c *Mumbai) GetGameBalance(ctx context.Context, user model.User, currency string, gameCode string, extra model.Extra) (balance int64, _err error) {
	// create the client to call the web-service.
	client, err := util.MumbaiFactory()
	if err != nil {
		return 0, err
	}

	username := os.Getenv("GAME_MUMBAI_MERCHANT_CODE") + os.Getenv("GAME_MUMBAI_AGENT_CODE") + fmt.Sprintf("%08s", user.IdAsString())
	_, err = c.LoginWithCreateUser(username, defaultPassword, extra.Ip, "")
	if err != nil {
		return 0, err
	}
	balanceFloat, err := client.CheckBalanceUser(username)
	if err != nil {
		return 0, ErrGetBalance
	}
	return util.MoneyInt(balanceFloat), nil
}
