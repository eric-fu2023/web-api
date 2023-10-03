package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
)

var (
	ErrEmptyCurrencyId      = errors.New("empty currency id")
	ErrFbCreateUserFailed   = errors.New("fb create user failed")
	ErrSabaCreateUserFailed = errors.New("saba create user failed")
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

	fbCurrency, fbCurrExists := currMap[consts.GameVendor["fb"]]
	if !fbCurrExists {
		return ErrEmptyCurrencyId
	}
	fbClient := util.FBFactory.NewClient()
	if res, e := fbClient.CreateUser(user.Username, []string{}, 0); e == nil {
		fbGpu := ploutos.GameVendorUser{
			ploutos.GameVendorUserC{
				GameVendorId:     consts.GameVendor["fb"],
				UserId:           user.ID,
				ExternalUserId:   user.Username,
				ExternalCurrency: fbCurrency,
				ExternalId:       fmt.Sprintf("%d", res),
			},
		}
		err = model.DB.Save(&fbGpu).Error
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFbCreateUserFailed, err)
		}
	}

	sabaCurrency, sabaCurrExists := currMap[consts.GameVendor["saba"]]
	if !sabaCurrExists {
		return ErrEmptyCurrencyId
	}
	sabaClient := util.SabaFactory.NewClient()
	if res, e := sabaClient.CreateMember(user.Username, sabaCurrency, os.Getenv("GAME_SABA_ODDS_TYPE")); e == nil {
		sabaGpu := ploutos.GameVendorUser{
			ploutos.GameVendorUserC{
				GameVendorId:     consts.GameVendor["saba"],
				UserId:           user.ID,
				ExternalUserId:   user.Username,
				ExternalCurrency: sabaCurrency,
				ExternalId:       res,
			},
		}
		err = model.DB.Save(&sabaGpu).Error
		if err != nil {
			return fmt.Errorf("%w: %w", ErrSabaCreateUserFailed, err)
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
