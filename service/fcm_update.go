package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

type FcmTokenUpdateService struct {
	FcmToken string `form:"fcm_token" json:"fcm_token"`
}

func (service *FcmTokenUpdateService) Update(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)

	u, _ := c.Get("user")
	user := u.(model.User)

	deviceInfo, err := util.GetDeviceInfo(c)
	if err != nil && errors.Is(err, util.ErrDeviceInfoEmpty) {
		return serializer.ParamErr(c, service, i18n.T("missing_device_uuid"), err)
	} else if err != nil {
		return serializer.ParamErr(c, service, i18n.T("invalid_device_info"), err)
	}
	if deviceInfo.Uuid == "" {
		return serializer.ParamErr(c, service, i18n.T("missing_device_uuid"), nil)
	}

	err = model.UpsertFcmToken(c, user.ID, deviceInfo.Uuid, service.FcmToken)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("UpsertFcmToken error: %s", err.Error())
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	return serializer.Response{
		Msg: i18n.T("success"),
	}
}
