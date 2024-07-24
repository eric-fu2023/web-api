package model

import (
	"context"
	"fmt"
	"time"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type CashMethod struct {
	ploutos.CashMethod

	CashMethodPromotion *ploutos.CashMethodPromotion `json:"cash_method,omitempty" form:"-" gorm:"references:CashMethodId;foreignKey:ID"`
}

func (CashMethod) GetByID(c *gin.Context, id int64, brandID int) (item CashMethod, err error) {
	err = DB.Where("brand_id = ? or brand_id = 0", brandID).Where("id", id).First(&item, id).Error
	return
}

func (CashMethod) List(c *gin.Context, withdrawOnly, topupOnly bool, platform string, brandID, vipID int) (list []CashMethod, err error) {
	u, _ := c.Get("user")
	user, _ := u.(User)

	var t []CashMethod
	q := DB.Preload("CashMethodChannel", "is_active").Where("is_active").Where("brand_id = ? or brand_id = 0", brandID)
	if withdrawOnly {
		q = q.Where("method_type < 0")
	}
	if topupOnly {
		q = q.Where("method_type > 0")
	}
	// var restrictPaymentChannel []int64 = user.RestrictPaymentChannel
	// if len(restrictPaymentChannel) != 0 {
	// 	q = q.Where("\"cash_methods\".id NOT IN ?", restrictPaymentChannel)
	// }

	now := time.Now().UTC()
	q = q.Joins("CashMethodPromotion", DB.Where("\"CashMethodPromotion\".start_at < ? and \"CashMethodPromotion\".end_at > ?", now, now).Where("\"CashMethodPromotion\".status = ?", 1).Where("\"CashMethodPromotion\".vip_id = ?", vipID))

	err = q.Order("sort desc").Find(&t).Error
	for i := range t {
		chns := FilterChannelByVip(c, user, t[i].CashMethodChannel)
		if len(chns) == 0 {
			continue
		}
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

func (CashMethod) ListWithAvailableChannel(c *gin.Context, withdrawOnly, topupOnly bool, platform string, brandID int) (list []CashMethod, err error) {
	var t []CashMethod
	q := DB.Preload("CashMethodChannel").Where("is_active").Where("brand_id = ? or brand_id = 0", brandID)
	if withdrawOnly {
		q = q.Where("method_type < 0")
	}
	if topupOnly {
		q = q.Where("method_type > 0")
	}
	err = q.Order("sort desc").Find(&t).Error
	for i := range t {
		if t[i].IsSupportedPlatform(platform) &&
			util.Reduce(t[i].CashMethodChannel, func(weight int64, ch ploutos.CashMethodChannel) int64 {
				return weight + ch.Weight
			}, 0) > 0 {
			list = append(list, t[i])
		}
	}
	return
}

func (CashMethod) GetByIDWithChannel(c *gin.Context, id int64) (item CashMethod, err error) {
	err = DB.Preload("CashMethodChannel.Stats").Where("id", id).First(&item, id).Error
	return
}

// success/failed/gateway_failed
func IncrementStats(stats models.CashMethodStats, result string) error {
	field := ""
	switch result {
	case "success":
		field = "success"
	case "failed":
		field = "failed"
	case "gateway_failed":
		field = "gateway_failed"
	}
	err := DB.Debug().Exec(fmt.Sprintf("update cash_method_stats set called = called + 1, %s = %s + 1 where id = ?", field, field), stats.ID).Error
	return err
}

func GetNextChannel(list []models.CashMethodChannel) models.CashMethodChannel {
	distribution := map[int64]int64{}
	var weightTotal int64 = 0
	accumulation := map[int64]int64{}
	var calledTotal int64 = 0
	for _, item := range list {
		distribution[item.ID] = item.Weight
		weightTotal += item.Weight
		accumulation[item.ID] = item.Stats.Called
		calledTotal += item.Stats.Called
	}
	for _, item := range list {
		if calledTotal == 0 {
			return item
		}
		if float64(accumulation[item.ID])/float64(calledTotal) < float64(distribution[item.ID])/float64(weightTotal) {
			return item
		}
	}
	return list[0]
}

func FilterChannelByVip(c context.Context, user User, chns []models.CashMethodChannel) []models.CashMethodChannel {
	ret := []models.CashMethodChannel{}
	vip, _ := GetVipWithDefault(c, user.ID)
	for _, ch := range chns {
		for _, lvl := range ch.VipLevels {
			if vip.VipRule.VIPLevel == int64(lvl) {
				ret = append(ret, ch)
				break
			}
		}
	}
	return ret
}

func FilterByAmount(c context.Context, amount int64, chns []models.CashMethodChannel) []models.CashMethodChannel {
	ret := []models.CashMethodChannel{}
	for _, ch := range chns {
		if amount <= ch.MaxAmount && amount >= ch.MinAmount {
			ret = append(ret, ch)
		}
	}
	return ret
}
