package service

import (
	"errors"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"

	"web-api/conf/consts"
	"web-api/model"
	"web-api/model/avatar"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserRegisterService struct {
	Username    string `form:"username" json:"username" binding:"required,username"`
	Password    string `form:"password" json:"password" binding:"required,password"`
	CurrencyId  int64  `form:"currency_id" json:"currency_id" binding:"required"`
	Code        string `form:"code" json:"code"`
	Channel     string `form:"channel" json:"channel"`
	Mobile      string `form:"mobile" json:"mobile"`
	CountryCode string `form:"country_code" json:"country_code"`

	Mutex *sync.Mutex
}

func validateMobileNumber(countryCode, mobile string) error {
	if countryCode == "+63" {
		phMobilepattern := `^(9\d{9})$`
		phMobileRegex := regexp.MustCompile(phMobilepattern)
		if !phMobileRegex.MatchString(mobile) {
			return errors.New("invalid mobile number format")
		}
	}
	return nil
}

func (service *UserRegisterService) validateMobileNumber() error {
	return validateMobileNumber(service.CountryCode, service.Mobile)
}

func (service *UserRegisterService) Register(c *gin.Context, bypassSetMobileOtpVerify bool) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	brandId := c.MustGet("_brand").(int)

	service.Username = strings.TrimSpace(strings.ToLower(service.Username))
	service.Code = strings.ToUpper(strings.TrimSpace(service.Code))

	service.CountryCode = util.FormatCountryCode(service.CountryCode)
	service.Mobile = strings.TrimPrefix(service.Mobile, "0")

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
			Channel:                 service.Channel, // replace with referrer's channel if referred

			Locale: c.MustGet("_locale").(string),
		},
	}

	if bypassSetMobileOtpVerify { // store mobile as unverified if flag is set
		if _err := service.validateMobileNumber(); _err != nil {
			return serializer.ParamErr(c, service, i18n.T("invalid_mobile_number_format"), _err)
		}
		unverifiedMobile, err := ploutos.ToEncrypt(service.Mobile)
		if err != nil {
			return serializer.ParamErr(c, service, i18n.T("invalid_mobile_number_format")+"encrypt", err)
		}
		var userWithMobile model.User
		uwmRows := model.DB.Where(`unverified_country_code`, service.CountryCode).Where(`unverified_mobile`, unverifiedMobile).First(&userWithMobile).RowsAffected
		if uwmRows > 0 {
			return serializer.ParamErr(c, service, i18n.T("existing_mobile"), nil)
		}

		user.UnverifiedCountryCode = service.CountryCode
		user.UnverifiedMobile = unverifiedMobile
	}

	genNickname(&user)
	user.Avatar = avatar.GetRandomAvatarUrl()

	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		err = CreateNewUserWithDB(&user, service.Code, tx)
		if err != nil {
			util.GetLoggerEntry(c).Errorf("CreateNewUser error: %s", err.Error())
			return
		}
		err = CreateUserWithDB(&user, tx)
		if err != nil {
			return
		}
		return
	})
	if err != nil {
		if errors.Is(err, ErrEmptyCurrencyId) {
			return serializer.ParamErr(c, service, i18n.T("empty_currency_id"), nil)
		}
		return serializer.DBErr(c, service, i18n.T("User_add_fail"), err)
	}

	tokenString, err := ProcessUserLogin(c, user, consts.AuthEventLoginMethod["username"], "", service.CountryCode, service.Mobile)
	if err != nil && errors.Is(err, ErrTokenGeneration) {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("Error_token_generation"), err)
	} else if err != nil && errors.Is(err, util.ErrInvalidDeviceInfo) {
		util.GetLoggerEntry(c).Errorf("processUserLogin error: %s", err.Error())
		return serializer.ParamErr(c, service, i18n.T("invalid_device_info"), err)
	} else if err != nil {
		util.GetLoggerEntry(c).Errorf("processUserLogin error: %s", err.Error())
		return serializer.GeneralErr(c, err)
	}

	go common.SendNotification(user.ID, consts.Notification_Type_User_Registration, i18n.T("notification_welcome_title"), i18n.T("notification_welcome"))

	return serializer.Response{
		Data: map[string]interface{}{
			"token": tokenString,
		},
	}
}
