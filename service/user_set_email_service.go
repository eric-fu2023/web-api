package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"github.com/gin-gonic/gin"
	"strings"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"
)

type UserSetEmailService struct {
	Email string `form:"email" json:"email" binding:"required"`
	Otp   string `form:"otp" json:"otp" binding:"required"`
}

func (service *UserSetEmailService) Set(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	user := c.MustGet("user").(model.User)
	service.Email = strings.ToLower(service.Email)

	userKeys := []string{service.Email}
	otp, err := cache.GetOtpByUserKeys(c, consts.SmsOtpActionSetEmail, userKeys)
	if err != nil && errors.Is(err, cache.ErrInvalidOtpAction) {
		return serializer.ParamErr(c, service, i18n.T("invalid_otp_action"), nil)
	}
	if err != nil {
		return serializer.GeneralErr(c, err)
	}
	if otp != service.Otp {
		return serializer.Err(c, service, serializer.CodeOtpInvalid, i18n.T("otp_invalid"), nil)
	}

	emailHash := serializer.MobileEmailHash(service.Email)
	var existing model.User
	rows := model.DB.Where(`email`, service.Email).Where(`email_hash`, emailHash).First(&existing).RowsAffected
	if rows > 0 {
		return serializer.ParamErr(c, service, i18n.T("existing_email"), nil)
	}

	user.Email = ploutos.EncryptedStr(service.Email)
	user.EmailHash = emailHash
	if err = model.DB.Updates(&user).Error; err != nil {
		return serializer.DBErr(c, service, i18n.T("email_update_failed"), err)
	}

	common.SendNotification(user.ID, consts.Notification_Type_Email_Reset, i18n.T("notification_email_add_title"), i18n.T("notification_email_add"))

	return serializer.Response{
		Msg: i18n.T("success"),
	}
}
