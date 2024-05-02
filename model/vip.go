package model

import (
	"context"
	"errors"
	"time"
	"web-api/cache"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/chenyahui/gin-cache/persist"
	"gorm.io/gorm"
)

const (
	VipRuleCacheKey = "vip_full_rule_cache"
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

func LoadVipRuleWithCache(c context.Context) (ret []models.VIPRule, err error) {
	err = cache.RedisStore.Get(VipRuleCacheKey, &ret)
	if errors.Is(err, persist.ErrCacheMiss) {
		err = DB.Where("is_active").Order("vip_level").Find(&ret).Error
		_ = cache.RedisStore.Set(VipRuleCacheKey, ret, 5*time.Minute)
	}
	return
}

func GetDefaultVip() (models.VIPRule, error) {
	var vipRule models.VIPRule
	err := DB.Where("is_active").Order("vip_level").First(&vipRule).Error
	return vipRule, err
}
