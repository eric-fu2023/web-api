package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type UserDeleteService struct {
	Otp string `form:"otp" json:"otp" binding:"required"`
}

func (service *UserDeleteService) Delete(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	otp := cache.RedisSessionClient.Get(context.TODO(), "otp:"+user.Email)
	if otp.Err() == redis.Nil {
		otp = cache.RedisSessionClient.Get(context.TODO(), "otp:"+user.CountryCode+user.Mobile)
	}
	if otp.Val() != service.Otp {
		return serializer.Err(c, service, serializer.CodeOtpInvalid, i18n.T("otp_invalid"), nil)
	}

	if rows := model.DB.Model(model.User{}).Where(`id`, user.ID).Update(`status`, 2).RowsAffected; rows < 1 {
		return serializer.DBErr(c, service, i18n.T("failed"), nil)
	}
	return serializer.Response{
		Msg: i18n.T("success"),
	}
}
