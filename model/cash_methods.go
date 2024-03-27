package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type CashMethod struct {
	ploutos.CashMethod
}

func (CashMethod) GetByID(c *gin.Context, id int64, brandID int) (item CashMethod, err error) {
	err = DB.Where("brand_id = ? or brand_id = 0", brandID).Where("id", id).First(&item, id).Error
	return
}

func (CashMethod) List(c *gin.Context, withdrawOnly, topupOnly bool, platform string, brandID int) (list []CashMethod, err error) {
	var t []CashMethod
	q := DB.Debug().Where("is_active").Where("brand_id = ? or brand_id = 0", brandID)
	if withdrawOnly {
		q = q.Where("method_type < 0")
	}
	if topupOnly {
		q = q.Where("method_type > 0")
	}
	err = q.Order("sort desc").Find(&t).Error
	for i := range t {
		if t[i].IsSupportedPlatform(platform) {
			list = append(list, t[i])
		}
	}
	return
}

func (a *CashMethod) IsSupportedPlatform(platform string) bool {
	switch platform {
	case "ios":
		return a.SupportIOS
	case "android":
		return a.SupportAndroid
	case "":
		return false
	default:
		return a.SupportWeb
	}
}
