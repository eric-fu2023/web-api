package service

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
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

	isAllowed := CheckRegistrationDeviceIPCount(deviceInfo.Uuid, c.ClientIP())
	if !isAllowed {
		return serializer.Err(c, service, serializer.CodeDBError, i18n.T("registration_restrict_exceed_count"), err)
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
	var referrer_nickname string
	var is_join_success bool

	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		ConnectChannelAgent(&user, tx)
		referrer_nickname, is_join_success, err = CreateNewUserWithDB(&user, service.Code, tx)
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

	// if register success, need to send to pixel
	if user.Channel == "pixel_app_001"{
		log.Printf("should log pixel event register for channel pixel_app_001")
		PixelRegisterEvent(user.ID, c.ClientIP(), os.Getenv("PIXEL_ACCESS_TOKEN"), os.Getenv("PIXEL_END_POINT"))
	}
	if user.Channel == "pixel_app_002"{
		log.Printf("should log pixel event register for channel pixel_app_002")
		PixelRegisterEvent(user.ID, c.ClientIP(), os.Getenv("PIXEL_ACCESS_TOKEN_002"), os.Getenv("PIXEL_END_POINT_002"))
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
			"token":            tokenString,
			"alliance_name":    referrer_nickname,
			"is_join_alliance": service.Code != "",
			"is_join_success":  is_join_success,
		},
	}
}

func ConnectChannelAgent(user *model.User, tx *gorm.DB) (err error) {
	// fmt.Printf("=== PRE username - %s, channelCode - %s, userChannelId - %s, userAgentId - %s === \n", user.Username, user.Channel, strconv.Itoa(int(user.ChannelId)), strconv.Itoa(int(user.AgentId)))
	// Default AgentId and ChannelId if no channelCode
	// Change to env later
	agentIdString := os.Getenv("DEFAULT_AGENT_ID")
	// channelCode := ""

	if agentIdString == "" || agentIdString == "1000000" {
		agentIdString = "1000001"
	}

	agentId, err := strconv.Atoi(agentIdString)

	if err != nil {
		return fmt.Errorf("string conv err: %w", err)
	}

	fmt.Printf("AgentId - %s \n", strconv.Itoa(int(agentId)))

	channel := ploutos.Channel{
		Code: user.Channel,
	}

	err = tx.Where(`code`, user.Channel).Find(&channel).Error
	if err != nil {
		return
	}

	if channel.ID != 0 {
		user.ChannelId = channel.ID
		user.AgentId = channel.AgentId
	} else {
		user.ChannelId = 1
		user.AgentId = int64(agentId)
	}

	// fmt.Printf("=== POST username - %s, channelCode - %s, userChannelId - %s, userAgentId - %s === \n", user.Username, user.Channel, strconv.Itoa(int(user.ChannelId)), strconv.Itoa(int(user.AgentId)))

	return
}

func CheckRegistrationDeviceIPCount(deviceId, ip string) (isAllowed bool) {
	registrationLoginRule, _ := model.GetRegistrationLoginRule()
	if registrationLoginRule.ID != 0 {
		deviceCount, ipCount := model.GetRegistrationDeviceIPCount(deviceId, ip)

		if deviceId != "" && registrationLoginRule.RegisterDeviceRestrict && deviceCount >= registrationLoginRule.RegisterDeviceRestrictCount {
			return false
		}

		if registrationLoginRule.RegisterIPRestrict && ipCount >= registrationLoginRule.RegisterIPRestrictCount {
			return false
		}
	}
	return true

}
