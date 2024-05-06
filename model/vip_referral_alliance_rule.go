package model

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
)

func GetAllReferralAllianceRules() (ret []models.VipReferralAllianceRule, err error) {
	err = DB.Model(&models.VipReferralAllianceRule{}).Order("game_category_id").Order("vip_level").Find(&ret).Error
	return
}
