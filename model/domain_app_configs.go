package model

import (
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

const (
	SupportTypeApp = "A"
	SupportTypeWeb = "W"
)

const (
	DomainTypeApi    = "A"
	DomainTypeRecord = "R"
	DomainTypeNami   = "M"
)

type DomainApiConfig struct {
	ploutos.DomainApiConfig
}

func (DomainApiConfig) FindDomainConfigs(supportType string, c *gin.Context) []DomainApiConfig {
	domains := []DomainApiConfig{}
	query := DB.Where("is_active")
	if len(supportType) > 0 {
		query.Where("support_type", supportType)
	}
	err := query.Find(&domains).Error
	if err != nil {
		util.GetLoggerEntry(c).Error("FindDomainsForApp failed: ", err.Error())
	}
	return domains
}
