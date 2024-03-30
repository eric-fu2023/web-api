package model

import (
	"context"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

func GetVip(c context.Context, userID int64) (ret models.VipRecord, err error) {
	err = DB.Preload("VipProgress").Preload("VipRule").Where("user_id", userID).First(&ret).Error
	return
}

func LoadRule(c context.Context) (ret []models.VIPRule, err error) {
	err = DB.Where("is_active").Order("vip_level").Find(&ret).Error
	return
}
