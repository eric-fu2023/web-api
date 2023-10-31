package service

import (
	"context"
	"regexp"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
)

type UserSecondaryPasswordService struct {
	SecondaryPassword string `form:"secondary_password" json:"secondary_password" binding:"required"`
	Otp               string `form:"otp" json:"otp" binding:"required"`
}

func (service *UserSecondaryPasswordService) SetSecondaryPassword(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	user := c.MustGet("user").(model.User)

	if matched, _ := regexp.MatchString(`^\d{6}$`, service.SecondaryPassword); !matched {
		return serializer.ParamErr(c, service, i18n.T("password_must_be_digits"), nil)
	}

	if user.Password != "" {
		if e := bcrypt.CompareHashAndPassword([]byte(user.SecondaryPassword), []byte(service.SecondaryPassword)); e == nil {
			return serializer.ParamErr(c, service, i18n.T("same_password"), nil)
		}
	}

	otp := cache.RedisSessionClient.Get(context.TODO(), "otp:"+user.Email)
	if otp.Err() == redis.Nil {
		otp = cache.RedisSessionClient.Get(context.TODO(), "otp:"+user.CountryCode+user.Mobile)
	}
	if otp.Val() != service.Otp {
		return serializer.ParamErr(c, service, i18n.T("验证码错误"), nil)
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(service.SecondaryPassword), model.PassWordCost)
	if err != nil {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("密码加密失败"), err)
	}

	if err = model.DB.Model(&user).Update(`secondary_password`, string(bytes)).Error; err != nil {
		return serializer.DBErr(c, service, i18n.T("密码修改失败"), err)
	}

	common.SendNotification(user, i18n.T("notification_pin_reset"))

	return serializer.Response{
		Msg: i18n.T("success"),
	}
}
