package model

import (
	"context"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

func LoadVipRebateRules(c context.Context) (ret []models.VipRebateRule, err error) {
	err = DB.Preload("GameVendorBrand.GameCategory").Order("game_vendor_id").Order("vip_level").Find(&ret).Error
	return
}
