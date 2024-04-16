package model

import (
	"context"
	"errors"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

func GetVipWithDefault(c context.Context, userID int64) (ret models.VipRecord, err error) {
	err = DB.Preload("VipProgress").Preload("VipRule").Where("user_id", userID).First(&ret).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		var rule models.VIPRule
		DB.Where("is_active").Order("vip_level").First(&rule)
		ret = models.VipRecord{
			UserID:  userID,
			VipRule: rule,
		}
		err = nil
	}
	return
}

func LoadVipRule(c context.Context) (ret []models.VIPRule, err error) {
	err = DB.Where("is_active").Order("vip_level").Find(&ret).Error
	return
}

func GetDefaultVip() (models.VIPRule, error) {
	var vipRule models.VIPRule
	err := DB.Where("is_active").Order("vip_level").First(&vipRule).Error
	return vipRule, err
}
