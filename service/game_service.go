package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/service/dc"
	"web-api/util/i18n"
)

var GameTypes = map[string]int64{
	"favourite": 1,
	"recent":    2,
}

type GameListService struct {
	common.Platform
	IsFeatured bool `form:"is_featured" json:"is_featured"`
}

func (service *GameListService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	brand := c.MustGet(`_brand`).(int)
	var games []ploutos.SubGameBrand
	if err = model.DB.Model(ploutos.SubGameBrand{}).Preload(`GameVendorBrand`).
		Scopes(model.ByGameIdsBrandAndIsFeatured([]string{}, int64(brand), service.IsFeatured), model.ByPlatformAndStatusOfSubAndVendor(service.Platform.Platform)).
		Find(&games).Error; err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var list []serializer.SubGameBrand
	for _, g := range games {
		list = append(list, serializer.BuildSubGameBrand(g))
	}

	r = serializer.Response{
		Data: list,
	}
	return
}

type UserRecentGameListService struct {
	common.Platform
}

func (service *UserRecentGameListService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	brand := c.MustGet(`_brand`).(int)
	u, _ := c.Get("user")
	user := u.(model.User)

	var games []ploutos.SubGameBrand
	var ids []string
	redisClient := cache.RedisRecentGamesClient
	re := redisClient.LRange(context.TODO(), fmt.Sprintf(`%s%d`, dc.RedisKeyRecentGames, user.ID), 0, -1)
	for _, v := range re.Val() {
		ids = append(ids, v)
	}

	if len(ids) > 0 {
		if err = model.DB.Model(ploutos.SubGameBrand{}).Preload(`GameVendorBrand`).
			Scopes(model.ByGameIdsBrandAndIsFeatured(ids, int64(brand), false), model.ByPlatformAndStatusOfSubAndVendor(service.Platform.Platform)).
			Find(&games).Error; err != nil {
			r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
			return
		}
	}

	var list []serializer.SubGameBrand
	for _, g := range games {
		list = append(list, serializer.BuildSubGameBrand(g))
	}

	r = serializer.Response{
		Data: list,
	}
	return
}
