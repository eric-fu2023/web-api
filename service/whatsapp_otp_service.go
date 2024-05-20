package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	whatsapputil "blgit.rfdev.tech/zhibo/utilities/whatsapp"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

type WhatsAppOtpService struct {
	CountryCode string `form:"country_code" json:"country_code" binding:"required"`
	Mobile      string `form:"mobile" json:"mobile" binding:"required,number"`
	CheckUser   bool   `form:"check_user" json:"check_user"`
	Action      string `form:"action" json:"action" binding:"required"`
	//common.Captcha
}

// TODO refactor along with sms_otp and email service, much of the logic is shared

func (service *WhatsAppOtpService) GetWhatsApp(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)

	service.CountryCode = util.FormatCountryCode(service.CountryCode)
	if service.Mobile[:1] == "0" {
		service.Mobile = service.Mobile[1:]
	}

	if service.CheckUser {
		exists := service.checkExisting(service.CountryCode, service.Mobile)
		if !exists {
			return serializer.ParamErr(c, service, i18n.T("account_invalid"), nil)
		}
	}

	err := service.verifyMobileNumber()
	if err != nil {
		return serializer.ParamErr(c, service, i18n.T("invalid_mobile_number_format"), nil)
	}

	otp, err := service.sendOtp(c, service.CountryCode+service.Mobile)
	if err != nil && errors.Is(err, cache.ErrInvalidOtpAction) {
		return serializer.ParamErr(c, service, i18n.T("invalid_otp_action"), nil)
	}
	if err != nil && errors.Is(err, errReachedOtpLimit) {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("otp_limit_reached"), err)
	}
	if err != nil && errors.Is(err, errOtpAlreadySent) {
		return serializer.Err(c, service, serializer.CodeSMSSent, i18n.T("Sms_wait"), nil)
	}
	if err != nil {
		util.GetLoggerEntry(c).Errorf("sendOtp error: %s", err.Error())
		return serializer.GeneralErr(c, err)
	}

	resp := serializer.SendOtp{}
	if os.Getenv("ENV") == "local" || os.Getenv("ENV") == "staging" {
		resp.Otp = otp
	}

	return serializer.Response{
		Msg:  i18n.T("success"),
		Data: resp,
	}
}

func (service *WhatsAppOtpService) sendOtp(c *gin.Context, otpUserKey string) (string, error) {
	// Check if otp has been sent
	otpSent, err := cache.GetOtp(c, service.Action, otpUserKey)
	if err != nil {
		return "", fmt.Errorf("GetOtp err: %w", err)
	}
	if otpSent != "" {
		return "", errOtpAlreadySent
	}

	// Generate new otp
	var otp string
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 6; i++ {
		otp += strconv.Itoa(rand.Intn(9))
	}

	// Send otp sms
	if os.Getenv("ENV") == "production" || os.Getenv("SEND_WHATSAPP_IN_TEST") == "true" {
		err = service.sendMessage(c, otp)
		if err != nil {
			return "", fmt.Errorf("sendMessage err: %w", err)
		}
	}

	// Set new otp in cache
	err = cache.SetOtp(c, service.Action, otpUserKey, otp)
	if err != nil {
		return "", fmt.Errorf("SetOtp err: %w", err)
	}

	return otp, nil
}

func (service *WhatsAppOtpService) sendMessage(c *gin.Context, otp string) error {
	// Check and increase OTP limit
	deviceInfo, _ := util.GetDeviceInfo(c)
	ip := c.ClientIP()
	isWithinLimit, err := cache.IncreaseSendOtpLimit(service.Mobile, ip, deviceInfo.Uuid, time.Now())
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Increase OTP limit error: %s", err.Error())
		return err
	}
	if !isWithinLimit {
		return errReachedOtpLimit
	}

	// Send WhatsApp Message
	cfg, err := whatsapputil.BuildDefaultConfig()
	if err != nil {
		util.GetLoggerEntry(c).Errorf("BuildDefaultConfig error: %s", err.Error())
		return err
	}

	manager := whatsapputil.Manager{
		Config: cfg,
	}

	res, err := manager.SendCode(service.CountryCode, service.Mobile, otp)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Send whatsapp message error: %s", err.Error())
	}
	if !res.HasSucceeded {
		return err
	}

	// Log OTP event
	event := model.OtpEvent{
		OtpEvent: ploutos.OtpEvent{
			CountryCode: service.CountryCode,
			Mobile:      service.Mobile,
			Otp:         otp,
			Provider:    whatsapputil.ProviderName[res.Provider],
			DateTime:    time.Now().Format(time.DateTime),
			BrandId:     int64(c.GetInt("_brand")),
			Method:      ploutos.OtpEventMethodWhatsApp,
		},
	}
	if err := model.LogOtpEvent(event); err != nil {
		// Just log error
		util.GetLoggerEntry(c).Errorf("log otp event error: %s", err.Error())
	}

	return nil
}

func (service *WhatsAppOtpService) verifyMobileNumber() error {
	if service.CountryCode == "+63" {
		phMobilepattern := `^(9\d{9})$`
		phMobileRegex := regexp.MustCompile(phMobilepattern)
		if !phMobileRegex.MatchString(service.Mobile) {
			return errors.New("invalid mobile number format")
		}
	}

	return nil
}

func (service *WhatsAppOtpService) checkExisting(countryCode, mobile string) bool {
	var user model.User
	mobileHash := util.MobileEmailHash(mobile)
	row := model.DB.Where(`country_code`, countryCode).Where(`mobile_hash`, mobileHash).First(&user).RowsAffected
	if row > 0 {
		return true
	}
	return false
}
