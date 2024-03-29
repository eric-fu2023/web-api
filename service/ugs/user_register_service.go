package ugs

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
	"web-api/model"
)

const IntegrationIdUGS = 1

type UGS struct {
}

func (c UGS) CreateWallet(user model.User, currency string) (err error) {
	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		//var isTestUser bool
		//if user.Role == 2 {
		//	isTestUser = true
		//}
		//client := util.UgsFactory.NewClient(cache.RedisClient)
		//_, _, err = client.GetUserToken(user.ID, user.Username, currency, lang, ip, isTestUser)
		//if err != nil {
		//	return
		//}
		var gameVendors []ploutos.GameVendor
		err = tx.Model(ploutos.GameVendor{}).Joins(`INNER JOIN game_vendor_brand gvb ON gvb.game_vendor_id = game_vendor.id AND gvb.brand_id = ?`, user.BrandId).
			Where(`game_vendor.game_integration_id`, IntegrationIdUGS).Find(&gameVendors).Error
		if err != nil {
			return
		}
		for _, gameVendor := range gameVendors {
			gvu := ploutos.GameVendorUser{
				GameVendorId:     gameVendor.ID,
				UserId:           user.ID,
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
