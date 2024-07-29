package model

import (
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func FindGeolocation(ip string, c *gin.Context) ploutos.Geolocation {
	geolocation := ploutos.Geolocation{}
	err := DB.Where("ip_address", ip).First(&geolocation).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		util.GetLoggerEntry(c).Warn("FindGeolocation failed: ", err.Error())
	}
	return geolocation
}
