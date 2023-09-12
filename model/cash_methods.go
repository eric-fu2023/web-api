package model

import (
	"web-api/conf/consts"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type CashMethod struct {
	models.CashMethodC
}

func (CashMethod) List(c *gin.Context, withdrawOnly, topupOnly bool, platform int64) (list []CashMethod, err error) {
	var t []CashMethod
	q := DB.Where("is_active")
	if withdrawOnly {
		q = q.Where("method_type < 0")
	}
	if topupOnly {
		q = q.Where("method_type > 0")
	}
	err = q.Order("sort desc").Find(&list).Error
	for i := range t {
		if t[i].IsSupportedPlatform(platform) {
			list = append(list, t[i])
		}
	}
	return
}

func (a *CashMethod) IsSupportedPlatform(platform int64) bool {
	switch consts.Platform[platform] {
	case "ios":
		return a.SupportIOS
	case "android":
		return a.SupportAndroid
	case "web":
		return a.SupportWeb
	default:
		return false
	}
}