package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"os"
	"strings"
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/service/dc"
	"web-api/service/dollar_jackpot"
	"web-api/service/fb"
	"web-api/service/imsb"
	"web-api/service/saba"
	"web-api/service/stream_game"
	"web-api/service/taya"
)

var (
	ErrEmptyCurrencyId           = errors.New("empty currency id")
	GameVendorUserRegisterStruct = map[string]common.UserRegisterInterface{
		"taya":           &taya.UserRegister{},
		"fb":             &fb.UserRegister{},
		"saba":           &saba.UserRegister{},
		"dc":             &dc.UserRegister{},
		"imsb":           &imsb.UserRegister{},
		"dollar_jackpot": &dollar_jackpot.UserRegister{},
		"stream_game":    &stream_game.UserRegister{},
	}
)

var (
	ErrTokenGeneration = errors.New("token generation error")
)

func CreateUser(user *model.User) (err error) {
	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Save(&user).Error
		if err != nil {
			return
		}

		userSum := ploutos.UserSum{
			UserId: user.ID,
		}
		tx2 := model.DB.Clauses(dbresolver.Use("txConn")).Begin()
		err = tx2.Error
		if err != nil {
			return
		}
		err = tx2.Create(&userSum).Error
		if err != nil {
			tx2.Rollback()
			return
		}

		userCounter := ploutos.UserCounter{
			UserId: user.ID,
		}
		err = tx.Create(&userCounter).Error
		if err != nil {
			tx2.Rollback()
			return
		}

		err = tx.Model(ploutos.User{}).Where(`id`, user.ID).Update(`setup_completed_at`, time.Now()).Error
		if err != nil {
			tx2.Rollback()
			return
		}

		var integrationCurrencies []ploutos.CurrencyGameIntegration
		err = tx.Where(`currency_id`, user.CurrencyId).Find(&integrationCurrencies).Error
		if err != nil {
			tx2.Rollback()
			return ErrEmptyCurrencyId
		}

		inteCurrMap := make(map[int64]string)
		for _, cur := range integrationCurrencies {
			inteCurrMap[cur.GameIntegrationId] = cur.Value
		}

		var gameIntegrations []ploutos.GameIntegration
		err = tx.Model(ploutos.GameIntegration{}).Find(&gameIntegrations).Error
		if err != nil {
			tx2.Rollback()
			return
		}
		for _, gi := range gameIntegrations {
			currency, exists := inteCurrMap[gi.ID]
			if !exists {
				tx2.Rollback()
				return ErrEmptyCurrencyId
			}
			err = common.GameIntegration[gi.ID].CreateWallet(*user, currency)
			if err != nil {
				tx2.Rollback()
				return
			}
		}

		// TODO: might remove in the future
		var currencies []ploutos.CurrencyGameVendor
		err = tx.Where(`currency_id`, user.CurrencyId).Find(&currencies).Error
		if err != nil {
			tx2.Rollback()
			return ErrEmptyCurrencyId
		}

		currMap := make(map[int64]string)
		for _, cur := range currencies {
			currMap[cur.GameVendorId] = cur.Value
		}

		games := strings.Split(os.Getenv("GAMES_REGISTERED_FOR_NEW_USER"), ",")
		for _, g := range games {
			if g == "" {
				continue
			}
			currency, exists := currMap[consts.GameVendor[g]]
			if !exists {
				tx2.Rollback()
				return ErrEmptyCurrencyId
			}
			game := GameVendorUserRegisterStruct[g]
			e := game.CreateUser(*user, currency)
			if e != nil && !errors.Is(e, game.VendorRegisterError()) { // if create vendor user failed, can proceed safely. when user first enter the game, it will retry
				tx2.Rollback()
				return fmt.Errorf("%w: %w", game.OthersError(), e)
			}
		}
		// TODO: END

		tx2.Commit()
		return
	})

	return
}

type MeService struct {
	WithKyc bool `form:"with_kyc" json:"with_kyc"`
}

func (service *MeService) Get(c *gin.Context) serializer.Response {
	u, _ := c.Get("user")
	user := u.(model.User)
	var userSum ploutos.UserSum
	if e := model.DB.Where(`user_id`, user.ID).First(&userSum).Error; e == nil {
		user.UserSum = &userSum
	}

	uaCond := model.GetUserAchievementCond{AchievementIds: []int64{
		model.UserAchievementIdFirstAppLoginTutorial,
		model.UserAchievementIdFirstAppLoginReward,
	}}
	userAchievements, err := model.GetUserAchievements(user.ID, uaCond)
	if err == nil {
		user.Achievements = userAchievements
	}

	if service.WithKyc {
		if value, err := GetCachedConfig(c, consts.ConfigKeyTopupKycCheck); err == nil {
			if value == "true" {
				user.KycCheckRequired = true
			}
		}
		var kyc model.Kyc
		if rows := model.DB.Scopes(model.ByUserId(user.ID), model.BySuccess).Order(`id DESC`).Find(&kyc).RowsAffected; rows > 0 {
			user.Kyc = &kyc
		}
	}
	return serializer.Response{
		Data: serializer.BuildUserInfo(c, user),
	}
}
