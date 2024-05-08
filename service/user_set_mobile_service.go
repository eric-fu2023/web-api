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
	"web-api/util"
	"web-api/util/i18n"
)

type UserSetMobileService struct {
	CountryCode string `form:"country_code" json:"country_code" binding:"required"`
	Mobile      string `form:"mobile" json:"mobile" binding:"required,number"`
	Otp         string `form:"otp" json:"otp" binding:"required"`
}

func (service *UserSetMobileService) Set(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	user := c.MustGet("user").(model.User)
	service.Mobile = strings.TrimPrefix(service.Mobile, "0")
	service.CountryCode = util.FormatCountryCode(service.CountryCode)

	userKeys := []string{service.CountryCode + service.Mobile}
	otp, err := cache.GetOtpByUserKeys(c, consts.SmsOtpActionSetMobile, userKeys)
	if err != nil && errors.Is(err, cache.ErrInvalidOtpAction) {
		return serializer.ParamErr(c, service, i18n.T("invalid_otp_action"), nil)
	}
	if err != nil {
		return serializer.GeneralErr(c, err)
	}
	if otp != service.Otp {
		return serializer.Err(c, service, serializer.CodeOtpInvalid, i18n.T("otp_invalid"), nil)
	}

	mobileHash := serializer.MobileEmailHash(service.Mobile)
	var existing model.User
	rows := model.DB.Where(`country_code`, service.CountryCode).Where(`mobile_hash`, mobileHash).First(&existing).RowsAffected
	if rows > 0 {
		return serializer.ParamErr(c, service, i18n.T("existing_mobile"), nil)
	}

	user.CountryCode = service.CountryCode
	user.Mobile = ploutos.EncryptedStr(service.Mobile)
	user.MobileHash = mobileHash
	if err = model.DB.Updates(&user).Error; err != nil {
		return serializer.DBErr(c, service, i18n.T("password_update_failed"), err)
	}

	common.SendNotification(user.ID, consts.Notification_Type_Mobile_Reset, i18n.T("notification_mobile_add_title"), i18n.T("notification_mobile_add"))

	return serializer.Response{
		Msg: i18n.T("success"),
	}
}
