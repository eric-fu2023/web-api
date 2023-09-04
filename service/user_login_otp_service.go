package service

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type UserLoginOtpService struct {
	CountryCode string `form:"country_code" json:"country_code"`
	Mobile      string `form:"mobile" json:"mobile"`
	Email       string `form:"email" json:"email"`
	Otp         string `form:"otp" json:"otp" binding:"required"`
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

	q :=  model.DB
	if service.CountryCode != "" && service.Mobile != "" {
		q = q.Where(`country_code = ? AND mobile = ?`, service.CountryCode, service.Mobile)
	} else {
		q = q.Where(`email`, service.Email)
	}
	setupRequired := false
	if rows := q.Find(&user).RowsAffected; rows == 0 { // new user
		user = model.User{
			models.UserC{
				Email:       service.Email,
				CountryCode: service.CountryCode,
				Mobile:      service.Mobile,
				Status:      1,
				Role:        1, // default role user
			},
		}
		user.BrandId = int64(c.MustGet("_brand").(int))
		user.AgentId = int64(c.MustGet("_agent").(int))
		err := model.DB.Create(&user).Error
		if err != nil {
			return serializer.DBErr(c, service, i18n.T("User_add_fail"), err)
		}
	}

	if user.Username == "" {
		setupRequired = true
	}

	tokenString, err := user.GenToken()
	if err != nil {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("Error_token_generation"), err)
	}
	cache.RedisSessionClient.Set(context.TODO(), user.GetRedisSessionKey(), tokenString, 20*time.Minute)

	return serializer.Response{
		Data: map[string]interface{}{
			"token": tokenString,
			"setup_required": setupRequired,
		},
	}
}
