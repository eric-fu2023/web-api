package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type UserSetPasswordService struct {
	Password string `form:"password" json:"password" binding:"required"`
	Otp      string `form:"otp" json:"otp" binding:"required"`
}

func (service *UserSetPasswordService) SetPassword(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	otp := cache.RedisSessionClient.Get(context.TODO(), "otp:"+user.Email)
	if otp.Err() == redis.Nil {
		otp = cache.RedisSessionClient.Get(context.TODO(), "otp:"+user.CountryCode+user.Mobile)
	}
	if otp.Val() != service.Otp {
		return serializer.ParamErr(i18n.T("验证码错误"), nil)
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(service.Password), model.PassWordCost)
	if err != nil {
		return serializer.ParamErr(i18n.T("密码加密失败"), err)
	}

	if err = model.DB.Model(&user).Update(`password`, string(bytes)).Error; err != nil {
		return serializer.ParamErr(i18n.T("密码修改失败"), err)
	}

	return serializer.Response{
		Msg: i18n.T("success"),
	}
}
