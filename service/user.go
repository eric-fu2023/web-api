package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"strings"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/service/dc"
	"web-api/service/fb"
	"web-api/service/saba"
)

var (
	ErrEmptyCurrencyId           = errors.New("empty currency id")
	GameVendorUserRegisterStruct = map[string]common.UserRegisterInterface{
		"fb":   &fb.UserRegister{},
		"saba": &saba.UserRegister{},
		"dc":   &dc.UserRegister{},
	}
)

func CreateUser(user model.User) error {
	tx := model.DB.Begin()
	err := tx.Save(&user).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	userSum := ploutos.UserSum{
		ploutos.UserSumC{
			UserId: user.ID,
		},
	}
	err = tx.Create(&userSum).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	var currencies []ploutos.CurrencyGameVendor
	err = model.DB.Where(`currency_id`, user.CurrencyId).Find(&currencies).Error
	if err != nil {
		return ErrEmptyCurrencyId
	}
	currMap := make(map[int64]string)
	for _, cur := range currencies {
		currMap[cur.GameVendorId] = cur.Value
	}

	games := strings.Split(os.Getenv("GAMES_REGISTERED_FOR_NEW_USER"), ",")
	for _, g := range games {
		currency, exists := currMap[consts.GameVendor[g]]
		if !exists {
			return ErrEmptyCurrencyId
		}
		game := GameVendorUserRegisterStruct[g]
		err = game.CreateUser(user, currency)
		if err != nil && !errors.Is(err, game.VendorRegisterError()) { // if create vendor user failed, can proceed safely. when user first enter the game, it will retry
			return fmt.Errorf("%w: %w", game.OthersError(), err)
		}
	}

	return nil
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
	if service.WithKyc {
		if value, err := GetCachedConfig(c, consts.ConfigKeyTopupKycCheck); err == nil {
			if value == "true" {
				user.KycCheckRequired = true
			}
		}
		var kyc model.Kyc
		if e := model.DB.Where(`user_id`, user.ID).Order(`id DESC`).First(&kyc).Error; e == nil {
			user.Kyc = &kyc
		}
	}
	return serializer.Response{
		Data: serializer.BuildUserInfo(c, user),
	}
}
