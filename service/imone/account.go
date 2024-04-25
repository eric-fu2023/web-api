package imone

import (
	"errors"

	"web-api/conf/consts"
	"web-api/model"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

func (c *ImOne) CreateUser(user model.User, currency string) error {
	gvu := ploutos.GameVendorUser{
		GameVendorId:     consts.GameVendor["imone"],
		UserId:           user.ID,
		ExternalUserId:   user.Username,
		ExternalCurrency: currency,
	}
	err := model.DB.Save(&gvu).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			err = c.VendorRegisterError()
		}
		return err
	}

	return util.ImOneFactory().CreateUser(user.IdAsString(), currency, user.Password, "")
}

// CreateWallet in db
// TODO check if wallets are created on imone user creation
func (c *ImOne) CreateWallet(user model.User, currency string) (err error) {
	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		var gameVendors []ploutos.GameVendor
		err = tx.Model(ploutos.GameVendor{}).Joins(`INNER JOIN game_vendor_brand gvb ON gvb.game_vendor_id = game_vendor.id`).
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
	if err != nil {
		return
	}
	return
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
