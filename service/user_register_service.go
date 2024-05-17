package service

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type UserRegisterService struct {
	Username   string `form:"username" json:"username" binding:"required,username"`
	Password   string `form:"password" json:"password" binding:"required,password"`
	CurrencyId int64  `form:"currency_id" json:"currency_id" binding:"required"`
	Code       string `form:"code" json:"code"`
	Channel    string `form:"channel" json:"channel"`
}

func (service *UserRegisterService) Register(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	brandId := c.MustGet("_brand").(int)
	service.Username = strings.TrimSpace(strings.ToLower(service.Username))
	service.Code = strings.ToUpper(strings.TrimSpace(service.Code))
	deviceInfo, err := util.GetDeviceInfo(c)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetDeviceInfo error: %s", err.Error())
		return serializer.ParamErr(c, service, i18n.T("invalid_device_info"), err)
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(service.Password), model.PassWordCost)
	if err != nil {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("password_encrypt_failed"), err)
	}
	var existing model.User
	rows := model.DB.Where(`username`, service.Username).First(&existing).RowsAffected
	if rows > 0 {
		return serializer.Err(c, service, serializer.CodeExistingUsername, i18n.T("existing_username"), nil)
	}
	var agent ploutos.Agent
	err = model.DB.Where(`brand_id`, brandId).Order(`id`).First(&agent).Error
	if err != nil {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
	}
	user := model.User{
		User: ploutos.User{
			Username:                service.Username,
			Password:                string(bytes),
			BrandId:                 int64(brandId),
			AgentId:                 agent.ID,
			CurrencyId:              service.CurrencyId,
			Status:                  1,
			Role:                    1, // default role user
			RegistrationIp:          c.ClientIP(),
			RegistrationDeviceUuid:  deviceInfo.Uuid,
			ReferralWagerMultiplier: 1,
			Channel:                 service.Channel,
		},
	}
	genNickname(&user)

	err = CreateNewUser(&user, service.Code)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("CreateNewUser error: %s", err.Error())
		return serializer.DBErr(c, service, i18n.T("User_add_fail"), err)
	}

	err = CreateUser(&user)
	if err != nil {
		if errors.Is(err, ErrEmptyCurrencyId) {
			return serializer.ParamErr(c, service, i18n.T("empty_currency_id"), nil)
		} else {
			return serializer.DBErr(c, service, i18n.T("User_add_fail"), err)
		}
	}

	tokenString, err := ProcessUserLogin(c, user, consts.AuthEventLoginMethod["username"], "", "", "")
	if err != nil && errors.Is(err, ErrTokenGeneration) {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("Error_token_generation"), err)
	} else if err != nil && errors.Is(err, util.ErrInvalidDeviceInfo) {
		util.GetLoggerEntry(c).Errorf("processUserLogin error: %s", err.Error())
		return serializer.ParamErr(c, service, i18n.T("invalid_device_info"), err)
	} else if err != nil {
		util.GetLoggerEntry(c).Errorf("processUserLogin error: %s", err.Error())
		return serializer.GeneralErr(c, err)
	}

	//go social_media_pixel.ReportRegisterConversion(c, user)
	go common.SendNotification(user.ID, consts.Notification_Type_User_Registration, i18n.T("notification_welcome_title"), i18n.T("notification_welcome"))

	return serializer.Response{
		Data: map[string]interface{}{
			"token": tokenString,
		},
	}
}
