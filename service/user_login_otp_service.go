package service

import (
	"blgit.rfdev.tech/taya/game-service/fb"
	"context"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"strings"
	"time"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type UserLoginOtpService struct {
	CountryCode string `form:"country_code" json:"country_code"`
	Mobile      string `form:"mobile" json:"mobile"`
	Email       string `form:"email" json:"email"`
	Otp         string `form:"otp" json:"otp" binding:"required"`
	Username    string `form:"username" json:"username"`
	Password    string `form:"password" json:"password"`
	CurrencyId  int64 `form:"currency_id" json:"currency_id" binding:"numeric"`
}

func (service *UserLoginOtpService) Login(c *gin.Context) serializer.Response {
	service.Email = strings.ToLower(service.Email)

	i18n := c.MustGet("i18n").(i18n.I18n)

	var user model.User
	errStr := ""
	key := "otp:"
	if service.Email != "" {
		key += service.Email
		errStr = i18n.T("Email_invalid")
	} else if service.CountryCode != "" && service.Mobile != "" {
		key += service.CountryCode + service.Mobile
		errStr = i18n.T("Mobile_invalid")
	} else {
		return serializer.ParamErr(c, service, i18n.T("Both_cannot_be_empty"), nil)
	}
	otp := cache.RedisSessionClient.Get(context.TODO(), key)
	if otp.Val() != service.Otp {
		return serializer.ParamErr(c, service, errStr, nil)
	}

	if rows := model.DB.Where(`email`, service.Email).Or(`country_code = ? AND mobile = ?`, service.CountryCode, service.Mobile).Find(&user).RowsAffected; rows == 0 { // new user
		if service.Username == "" || service.Password == "" {
			return serializer.ParamErr(c, service, i18n.T("empty_username_password"), nil)
		}
		if service.CurrencyId == 0 {
			return serializer.ParamErr(c, service, i18n.T("empty_currency_id"), nil)
		}
		var existing model.User
		if r := model.DB.Where(`username`, service.Username).Limit(1).Find(&existing).RowsAffected; r != 0 {
			return serializer.Err(c, service, serializer.CodeExistingUsername, i18n.T("existing_username"), nil)
		}
		bytes, err := bcrypt.GenerateFromPassword([]byte(service.Password), model.PassWordCost)
		if err != nil {
			return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("密码加密失败"), err)
		}
		tx := model.DB.Begin()
		user = model.User{
			Email:       service.Email,
			CountryCode: service.CountryCode,
			Mobile:      service.Mobile,
			Username:    service.Username,
			Password:    string(bytes),
			Status:      1,
			Role:        1, // default role user
		}
		user.BrandId = int64(c.MustGet("_brand").(int))
		user.AgentId = int64(c.MustGet("_agent").(int))
		err = tx.Create(&user).Error
		if err != nil {
			tx.Rollback()
			return serializer.DBErr(c, service, i18n.T("User_add_fail"), err)
		}

		var currency model.CurrencyGameProvider
		err = model.DB.Where(`game_provider_id`, consts.GameProvider["fb"]).Where(`currency_id`, service.CurrencyId).First(&currency).Error
		if err != nil {
			tx.Rollback()
			return serializer.ParamErr(c, service, i18n.T("empty_currency_id"), nil)
		}
		client := fb.FB{
			MerchantId: "1552945083054354433",
			MerchantApiSecret: "Lc63hMKwQz0R8Y4MbB7F6mhCbzLuZoU9",
			IsSandbox: true,
		}
		res, err := client.CreateUserAndWallet(user.Username, []int64{currency.Value}, 0)
		if err != nil {
			tx.Rollback()
			return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("fb_create_user_failed"), err)
		}
		if externalUserId, ok := res.(float64); ok {
			gpu := model.GameProviderUser{
				GameProviderId: consts.GameProvider["fb"],
				UserId: user.ID,
				ExternalUserId: strconv.Itoa(int(externalUserId)),
				CurrencyGameProviderId: currency.ID,
			}
			err = tx.Save(&gpu).Error
			if err != nil {
				tx.Rollback()
				return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("fb_create_user_failed"), err)
			}
		}
		tx.Commit()
	}

	tokenString, err := user.GenToken()
	if err != nil {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("Error_token_generation"), err)
	}
	cache.RedisSessionClient.Set(context.TODO(), strconv.Itoa(int(user.ID)), tokenString, 20*time.Minute)

	return serializer.Response{
		Data: map[string]interface{}{
			"token": tokenString,
		},
	}
}
