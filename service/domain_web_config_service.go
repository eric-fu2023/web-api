package service

import (
	"web-api/model"

	"github.com/gin-gonic/gin"
)

const (
	RedisKeyDomainWebConfigs = "domain_web_configs:"
)

type DomainWebConfigService struct {
}

func (service *DomainWebConfigService) RetrieveChannelForOrigin(c *gin.Context) string {
	var origin string
	if origin = c.Request.Header.Get("ori"); len(origin) < 1 {
		return ""
	}
	domain := model.DomainWebConfig{}.FindDomainWebConfig(c, origin)
	if domain != nil && domain.ID > 0 {
		return domain.Channel
	}
	return ""
}
