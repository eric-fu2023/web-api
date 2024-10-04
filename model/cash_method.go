package model

import (
	"context"
	"time"

	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type CashMethod struct {
	ploutos.CashMethod

	CashMethodChannels  []ploutos.CashMethodChannel  `gorm:"foreignKey:CashMethodId;references:ID"`
	CashMethodPromotion *ploutos.CashMethodPromotion `json:"cash_method_promotion,omitempty" form:"-" gorm:"foreignKey:CashMethodId;references:ID"`
}

func (CashMethod) GetByID(c *gin.Context, id int64, brandID int) (item CashMethod, err error) {
	err = DB.Where("brand_id = ? or brand_id = 0", brandID).Where("id", id).First(&item, id).Error
	return
}

func (CashMethod) GetByIDWithChannel(c *gin.Context, id int64) (item CashMethod, err error) {
	err = DB.Preload("CashMethodChannels", "is_active").Preload("CashMethodChannels.Stats").Where("id", id).First(&item, id).Error
	return
}

func (CashMethod) List(c context.Context, withdrawOnly, topupOnly bool, platform string, brandId, vipId int, user User) ([]CashMethod, error) {
	var cashMethods []CashMethod
	q := DB.Debug().Preload("CashMethodChannels", "is_active").Where("is_active").Where("brand_id = ? or brand_id = 0", brandId)
	if withdrawOnly {
		q = q.Where("method_type < 0")
	}
	if topupOnly {
		q = q.Where("method_type > 0")
	}
	var restrictPaymentChannel []int64 = user.RestrictPaymentChannel
	if len(restrictPaymentChannel) != 0 {
		q = q.Where("\"cash_methods\".id NOT IN ?", restrictPaymentChannel)
	}

	now := time.Now().UTC()
	q = q.Joins("CashMethodPromotion", DB.Where("\"CashMethodPromotion\".start_at < ? and \"CashMethodPromotion\".end_at > ?", now, now).Where("\"CashMethodPromotion\".status = ?", 1).Where("\"CashMethodPromotion\".vip_id = ?", vipId))

	err := q.Order("sort desc").Find(&cashMethods).Error
	if err != nil {
		return []CashMethod{}, err
	}

	var filteredCashMethods []CashMethod
	for i := range cashMethods {
		chns := FilterCashMethodChannelsByVip(c, user, cashMethods[i].CashMethodChannels)
		if len(chns) == 0 {
			continue
		}
		if cashMethods[i].IsSupportedPlatform(platform) {
			filteredCashMethods = append(filteredCashMethods, cashMethods[i])
		}
	}
	return filteredCashMethods, err
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

func (CashMethod) ListWithAvailableChannel(c *gin.Context, withdrawOnly, topupOnly bool, platform string, brandID int) (list []CashMethod, err error) {
	var t []CashMethod
	q := DB.Preload("CashMethodChannels").Where("is_active").Where("brand_id = ? or brand_id = 0", brandID)
	if withdrawOnly {
		q = q.Where("method_type < 0")
	}
	if topupOnly {
		q = q.Where("method_type > 0")
	}
	err = q.Order("sort desc").Find(&t).Error
	for i := range t {
		if t[i].IsSupportedPlatform(platform) &&
			util.Reduce(t[i].CashMethodChannels, func(weight int64, ch ploutos.CashMethodChannel) int64 {
				return weight + ch.Weight
			}, 0) > 0 {
			list = append(list, t[i])
		}
	}
	return
}
