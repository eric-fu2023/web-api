package service

import (
	"context"
	"fmt"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/service/dc"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
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

type GameByStreamerService struct {
	common.Platform
	StreamerId int64 `form:"streamer_id" json:"streamer_id"`
}

func (service *GameByStreamerService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var game_id int64
	err=model.DB.
	Select("stream_game_id").
	Table("stream_game_users").
	Where("user_id = ?", service.StreamerId).
	Where("game_type = ?", consts.ExternalGame).
	Where("deleted_at is null").
	Order("created_at desc").
	Limit(1).
	Find(&game_id).Error
	if err!=nil{
		fmt.Println("get game_id in stream_gae_user failed, ", err)
		return
	}
	var game ploutos.SubGameBrand
	if err = model.DB.Model(ploutos.SubGameBrand{}).Preload(`GameVendorBrand`).
		Where("id", game_id).Find(&game).Error; err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	r = serializer.Response{
		Data: game,
	}
	return
}
