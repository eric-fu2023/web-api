package model

import (
	"time"
	"web-api/cache"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	RedisKeyDomainWebConfigs        = "domain_web_configs:"
	RedisExpirationDomainWebConfigs = 3 * 60 * time.Second
)

type DomainWebConfig struct {
	ploutos.DomainWebConfig

	RedirectTos []DomainWebConfig `gorm:"references:ID;foreignKey:RedirectFromId"`
}

func (DomainWebConfig) FindByOrigin(c *gin.Context, origin string) (domain *DomainWebConfig) {
	// retrieve from Redis
	if util.FindFromRedis(c, cache.RedisDomainConfigClient, RedisKeyDomainWebConfigs+origin, &domain); domain != nil {
		return
	}
	// retrieve from DB
	if err := DB.Preload("RedirectTos", "is_active").Where("is_active").Where("origin", origin).Find(&domain).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			util.GetLoggerEntry(c).Warn("FindDomainsForApp failed: ", err.Error())
		}
		return
	}
	// cache in Redis
	if domain != nil {
		util.CacheIntoRedis(c, cache.RedisDomainConfigClient, RedisKeyDomainWebConfigs+origin, RedisExpirationDomainWebConfigs, domain)
	}
	return
}
