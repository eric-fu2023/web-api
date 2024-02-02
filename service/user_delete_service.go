package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"web-api/cache"
	"web-api/conf/consts"
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

	userKeys := []string{user.Email, user.CountryCode + user.Mobile}
	otp, err := cache.GetOtpByUserKeys(c, consts.SmsOtpActionDeleteUser, userKeys)
	if err != nil && errors.Is(err, cache.ErrInvalidOtpAction) {
		return serializer.ParamErr(c, service, i18n.T("invalid_otp_action"), nil)
	}
	if err != nil {
		return serializer.GeneralErr(c, err)
	}

	if otp != service.Otp {
		return serializer.Err(c, service, serializer.CodeOtpInvalid, i18n.T("otp_invalid"), nil)
	}

	if rows := model.DB.Model(model.User{}).Where(`id`, user.ID).Update(`status`, 2).RowsAffected; rows < 1 {
		return serializer.DBErr(c, service, i18n.T("failed"), nil)
	}
	return serializer.Response{
		Msg: i18n.T("success"),
	}
}
