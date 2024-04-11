package game_integration

import (
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GetUrlService struct {
	GameId   int64 `form:"game_id" json:"game_id" binding:"required"`
	Platform int64 `form:"platform" json:"platform" binding:"required"`
}

func (service *GetUrlService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	lang := c.MustGet("_locale").(string)
	user := c.MustGet("user").(model.User)

	var subGame ploutos.SubGameC
	err = model.DB.Preload(`GameVendor`).Where(`id`, service.GameId).First(&subGame).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var gvu ploutos.GameVendorUser
	err = model.DB.Where(`user_id`, user.ID).Where(`game_vendor_id`, subGame.VendorId).First(&gvu).Error
	if err != nil {
		return
	}

	var url string
	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		var lastPlayed ploutos.GameVendorUser
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload(`GameVendor`).Where(`user_id`, user.ID).Where(`is_last_played`, true).
			Order(`updated_at DESC`).Limit(1).Find(&lastPlayed).Error
		if err != nil {
			return
		}
		if lastPlayed.ID != 0 && lastPlayed.GameVendorId != int64(subGame.VendorId) { // transfer out from the game is needed
			gameFrom := common.GameIntegration[lastPlayed.GameVendor.GameIntegrationId]
			err = gameFrom.TransferFrom(tx, user, lastPlayed.ExternalCurrency, lang, lastPlayed.GameVendor.GameCode, c.ClientIP())
			if err != nil {
				return
			}
			err = tx.Model(ploutos.GameVendorUser{}).Where(`id`, lastPlayed.ID).Updates(map[string]interface{}{"balance": 0, "is_last_played": false}).Error
			if err != nil {
				return
			}
		}
		var sum ploutos.UserSum
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(`user_id`, user.ID).First(&sum).Error
		if err != nil {
			return
		}
		game := common.GameIntegration[subGame.GameVendor.GameIntegrationId]
		var transferToBalance int64
		if sum.Balance > 0 { // transfer in to the game is needed
			transferToBalance, err = game.TransferTo(tx, user, sum, gvu.ExternalCurrency, lang, subGame.GameVendor.GameCode, c.ClientIP())
			if err != nil {
				return
			}
		}
		url, err = game.GetGameUrl(user, gvu.ExternalCurrency, lang, subGame.GameVendor.GameCode, subGame.GameCode, c.ClientIP(), service.Platform)
		if err != nil {
			return
		}
		err = tx.Model(ploutos.GameVendorUser{}).Where(`game_vendor_id`, subGame.GameVendor.ID).Where(`user_id`, user.ID).Updates(map[string]interface{}{"balance": transferToBalance, "is_last_played": true}).Error
		if err != nil {
			return
		}
		return
	})

	r = serializer.Response{
		Data: url,
	}
	return
}

type GameCategoryListService struct {
}

func (service *GameCategoryListService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var categories []ploutos.GameCategory

	if err = model.DB.Model(ploutos.GameCategory{}).Preload(`GameVendor`, "game_integration_id = 1").
		Find(&categories).Error; err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	var data []serializer.GameCategory
	for _, cat := range categories {
		var subGameIds []int64
		var gameId int64
		if len(cat.GameVendor) > 0 {
			for _, v := range cat.GameVendor {
				model.DB.Model(ploutos.SubGameC{}).Select("id").Where("vendor_id = ?", v.ID).Where("game_code = ?", "lobby").Find(&gameId)
				subGameIds = append(subGameIds, gameId)
			}
		}
		data = append(data, serializer.BuildGameCategory(c, cat, subGameIds))
	}
	r = serializer.Response{
		Data: data,
	}
	return
}
