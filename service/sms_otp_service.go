package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	smsutil "blgit.rfdev.tech/zhibo/utilities/sms"
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

var (
	errIgnoreCountry   = errors.New("ignore country")
	errReachedOtpLimit = errors.New("reached otp limit")
	errOtpAlreadySent  = errors.New("otp has already been sent")
)

type SmsOtpService struct {
	CountryCode string `form:"country_code" json:"country_code" binding:"required,startswith=+"`
	Mobile      string `form:"mobile" json:"mobile" binding:"required,number"`
	CheckUser   bool   `form:"check_user" json:"check_user"`
	Action      string `form:"action" json:"action" binding:"required"`
	//common.Captcha
}

func (service *SmsOtpService) GetSMS(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)

	//err := common.CheckCaptcha(service.PointJson, service.Token)
	//if err != nil {
	//	return serializer.Err(c, service, serializer.CodeCaptchaInvalid, i18n.T("invalid_captcha"), nil)
	//}

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
	if err != nil && errors.Is(err, errIgnoreCountry) {
		return serializer.Response{
			Msg: i18n.T("success"),
		}
	} else if err != nil {
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

func (service *SmsOtpService) GetUsernameSMS(c *gin.Context, username string) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)

	otp, err := service.sendOtp(c, username)
	if err != nil && errors.Is(err, cache.ErrInvalidOtpAction) {
		return serializer.ParamErr(c, service, i18n.T("invalid_otp_action"), nil)
	}
	if err != nil && errors.Is(err, errReachedOtpLimit) {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("otp_limit_reached"), err)
	}
	if err != nil && errors.Is(err, errOtpAlreadySent) {
		return serializer.Err(c, service, serializer.CodeSMSSent, i18n.T("Sms_wait"), nil)
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

func (service *SmsOtpService) sendOtp(c *gin.Context, otpUserKey string) (string, error) {
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
	if os.Getenv("ENV") == "production" || os.Getenv("SEND_SMS_IN_TEST") == "true" {
		err = service.sendSMS(c, otp)
		if err != nil {
			return "", fmt.Errorf("sendSMS err: %w", err)
		}
	}

	// Set new otp in cache
	err = cache.SetOtp(c, service.Action, otpUserKey, otp)
	if err != nil {
		return "", fmt.Errorf("SetOtp err: %w", err)
	}

	return otp, nil
}

func (service *SmsOtpService) sendSMS(c *gin.Context, otp string) error {
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

	// Send SMS
	smsManager := smsutil.Manager{
		Templates: util.BuildSmsTemplates(c),
		Config:    smsutil.BuildDefaultConfig(),
	}
	res, err := smsManager.Send(service.CountryCode, service.Mobile, otp)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Send sms error: %s", err.Error())
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
			Provider:    smsutil.SmsProviderName[res.Provider],
			DateTime:    time.Now().Format(time.DateTime),
			BrandId:     int64(c.GetInt("_brand")),
		},
	}
	if err := model.LogOtpEvent(event); err != nil {
		// Just log error
		util.GetLoggerEntry(c).Errorf("log otp event error: %s", err.Error())
	}

	return nil
}

func (service *SmsOtpService) verifyMobileNumber() error {
	if service.CountryCode != "+63" && service.CountryCode != "+65" {
		return errIgnoreCountry
	}

	if service.CountryCode == "+63" {
		phMobilepattern := `^(9\d{9})$`
		phMobileRegex := regexp.MustCompile(phMobilepattern)
		if !phMobileRegex.MatchString(service.Mobile) {
			return errors.New("invalid mobile number format")
		}
	}

	return nil
}

func (service *SmsOtpService) checkExisting(countryCode, mobile string) bool {
	var user model.User
	row := model.DB.Where(`country_code`, countryCode).Where(`mobile`, mobile).First(&user).RowsAffected
	if row > 0 {
		return true
	}
	return false
}
