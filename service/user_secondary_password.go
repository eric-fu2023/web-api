package service

import (
	"errors"
	"regexp"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"

	"github.com/gin-gonic/gin"
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

	userKeys := []string{string(user.Email), user.CountryCode + string(user.Mobile)}
	otp, err := cache.GetOtpByUserKeys(c, consts.SmsOtpActionSetSecondaryPassword, userKeys)
	if err != nil && errors.Is(err, cache.ErrInvalidOtpAction) {
		return serializer.ParamErr(c, service, i18n.T("invalid_otp_action"), nil)
	}
	if err != nil {
		return serializer.GeneralErr(c, err)
	}

	if otp != service.Otp {
		if user.Role != 2 || (user.Role == 2 && service.Otp != "159357"){
			return serializer.Err(c, service, serializer.CodeOtpInvalid, i18n.T("otp_invalid"), nil)
		}
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(service.SecondaryPassword), model.PassWordCost)
	if err != nil {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("password_encrypt_failed"), err)
	}

	if err = model.DB.Model(&user).Update(`secondary_password`, string(bytes)).Error; err != nil {
		return serializer.DBErr(c, service, i18n.T("pin_update_failed"), err)
	}

	common.SendNotification(user.ID, consts.Notification_Type_Pin_Reset, i18n.T("notification_pin_reset_title"), i18n.T("notification_pin_reset"))

	return serializer.Response{
		Msg: i18n.T("success"),
	}
}
