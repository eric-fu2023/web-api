package game_integration

import (
	"fmt"
	"log"

	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GetUrlService struct {
	SubGameId int64 `form:"game_id" json:"game_id" binding:"required"`
	Platform  int64 `form:"platform" json:"platform" binding:"required"`
}

func (service *GetUrlService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	locale := c.MustGet("_locale").(string)
	user := c.MustGet("user").(model.User)

	var subGame ploutos.SubGameC
	err = model.DB.Preload(`GameVendor`).Where(`id`, service.SubGameId).First(&subGame).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var gvu ploutos.GameVendorUser
	err = model.DB.Where(`user_id`, user.ID).Where(`game_vendor_id`, subGame.VendorId).First(&gvu).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	game := common.GameIntegration[subGame.GameVendor.GameIntegrationId]
	extra := model.Extra{Locale: locale, Ip: c.ClientIP()}
	url, err := game.GetGameUrl(user, gvu.ExternalCurrency, subGame.GameVendor.GameCode, subGame.GameCode, service.Platform, extra)
	log.Printf("game.GetGameUrl url:%s err: %v user %v, gvu.ExternalCurrency %v, subGame.GameVendor.GameCode %v, subGame.GameCode %v, service.Platform %v, extra %+v", url, err, user, gvu.ExternalCurrency, subGame.GameVendor.GameCode, subGame.GameCode, service.Platform, extra)
	if err != nil {
		return
	}
	go func(user model.User, lang string, subGame ploutos.SubGameC, game common.GameIntegrationInterface, gvu ploutos.GameVendorUser) {
		model.GlobalWaitGroup.Add(1)
		defer model.GlobalWaitGroup.Done()

		err = model.DB.Transaction(func(_tx *gorm.DB) (err error) {
			var lastPlayed ploutos.GameVendorUser
			err = _tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload(`GameVendor`).Where(`user_id`, user.ID).Where(`is_last_played`, true).
				Order(`updated_at DESC`).Limit(1).Find(&lastPlayed).Error
			if err != nil {
				return
			}

			if lastPlayed.ID != 0 && lastPlayed.GameVendorId != int64(subGame.VendorId) { // transfer out from the game is needed
				gameFrom := common.GameIntegration[lastPlayed.GameVendor.GameIntegrationId]
				err = gameFrom.TransferFrom(_tx, user, lastPlayed.ExternalCurrency, lastPlayed.GameVendor.GameCode, lastPlayed.GameVendorId, extra)
				if err != nil {
					return
				}
				err = _tx.Model(ploutos.GameVendorUser{}).Where(`id`, lastPlayed.ID).Updates(map[string]interface{}{"balance": 0, "is_last_played": false}).Error
				if err != nil {
					return
				}
			}
			return
		})
		if err != nil {
			util.Log().Error(`GAME INTEGRATION TRANSFER OUT ERROR: %v`, err)
			return
		} else {
			util.Log().Info(`GAME INTEGRATION TRANSFER OUT OK`)
		}

		err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
			var sum ploutos.UserSum
			err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(`user_id`, user.ID).First(&sum).Error
			if err != nil {
				return
			}

			var transferToBalance int64
			if sum.Balance > 0 { // transfer in to the game is needed
				transferToBalance, err = game.TransferTo(tx, user, sum, gvu.ExternalCurrency, subGame.GameVendor.GameCode, subGame.GameVendor.ID, extra)
				if err != nil {
					return
				}
			}

			err = tx.Model(ploutos.GameVendorUser{}).Where(`game_vendor_id`, subGame.GameVendor.ID).Where(`user_id`, user.ID).Updates(map[string]interface{}{"balance": gorm.Expr(`balance + ?`, transferToBalance), "is_last_played": true}).Error
			if err != nil {
				return
			}
			err = tx.Model(ploutos.UserSum{}).Where(`user_id`, user.ID).Update("is_recall_needed", true).Error
			return err
		})
		if err != nil {
			util.Log().Error(`GAME INTEGRATION TRANSFER IN ERROR user_id: %d, game_code: %s, %v`, user.ID, subGame.GameVendor.GameCode, err)
		}
		defer func() {
			go func() {
				userSum, _ := model.GetByUserIDWithLockWithDB(user.ID, model.DB)
				common.SendUserSumSocketMsg(user.ID, userSum.UserSum, "enter_game", 0)
			}()
		}()
	}(user, locale, subGame, game, gvu)

	r = serializer.Response{
		Data: url,
	}
	return
}

type GameCategoryListService struct {
	Platform int64 `form:"platform" json:"platform" binding:"required"`
}

func (service *GameCategoryListService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var categories []ploutos.GameCategory
	platform, ok := consts.PlatformIdToGameVendorColumn[service.Platform]
	if !ok {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("invalid_platform"), err)
		return
	}

	if err = model.DB.Model(ploutos.GameCategory{}).Preload(`GameVendorBrand`, fmt.Sprintf("status = 1 AND %s = 1", platform)).Preload(`GameVendorBrand.GameVendor`, "game_integration_id = 1").
		Find(&categories).Error; err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	var data []serializer.GameCategory
	for _, cat := range categories {
		var subGameIds []int64
		var gameId int64
		if len(cat.GameVendorBrand) > 0 {
			for _, v := range cat.GameVendorBrand {
				// temporary hardcode, will change later
				model.DB.Model(ploutos.SubGameC{}).Select("id").Where("vendor_id = ?", v.GameVendorId).Where("game_code = ? OR (game_code = ? AND vendor_id = 11) OR (game_code = ? AND vendor_id = 12) OR (game_code = ? AND vendor_id = 13)", "lobby", "200", "8", "0").Find(&gameId)
				subGameIds = append(subGameIds, gameId)
			}
		}

		gameCategory := serializer.BuildGameCategory(c, cat, subGameIds)
		// catering for frontend to not return categories without vendors & sports category
		if gameCategory.Id == 0 || gameCategory.Id == 1 {
			continue
		}
		data = append(data, gameCategory)
	}
	r = serializer.Response{
		Data: data,
	}
	return
}
