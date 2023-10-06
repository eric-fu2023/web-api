package service

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"blgit.rfdev.tech/zhibo/utilities"
	"context"
	"errors"
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
	errIgnoreCountry = errors.New("ignore country")
)

type SmsOtpService struct {
	CountryCode string `form:"country_code" json:"country_code" binding:"required,startswith=+"`
	Mobile      string `form:"mobile" json:"mobile" binding:"required,number"`
}

func (service *SmsOtpService) GetSMS(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)

	if service.Mobile[:1] == "0" {
		service.Mobile = service.Mobile[1:]
	}

	err := service.verifyMobileNumber()
	if err != nil && errors.Is(err, errIgnoreCountry) {
		return serializer.Response{
			Data: serializer.User{
				CountryCode: service.CountryCode,
				Mobile:      service.Mobile,
			},
		}
	} else if err != nil {
		return serializer.ParamErr(c, service, i18n.T("invalid_mobile_number_format"), nil)
	}

	otpSent := cache.RedisSessionClient.Get(context.TODO(), "otp:"+service.CountryCode+service.Mobile)
	if otpSent.Val() != "" {
		return serializer.ParamErr(c, service, i18n.T("Sms_wait"), nil)
	}

	var otp string
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 6; i++ {
		otp += strconv.Itoa(rand.Intn(9))
	}

	if os.Getenv("ENV") == "production" || os.Getenv("SEND_SMS_IN_TEST") == "true" {
		if err := service.sendSMS(c, otp); err != nil {
			return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("Sms_fail"), err)
		}
	}
	cache.RedisSessionClient.Set(context.TODO(), "otp:"+service.CountryCode+service.Mobile, otp, 2*time.Minute)

	msg := i18n.T("success")
	if os.Getenv("ENV") == "local" || os.Getenv("ENV") == "staging" {
		msg = otp
	}

	return serializer.Response{
		Msg: msg,
	}
}

func (service *SmsOtpService) sendSMS(c *gin.Context, otp string) error {
	i18n := c.MustGet("i18n").(i18n.I18n)

	smsManager := utilities.SmsManager{
		HuanXunTemplate: `您的验证码是 %s，5分钟有效，请尽快验证`,
		BulkSmsTemplate: i18n.T("Your_request_otp"),
		AwsSnsTemplate:  i18n.T("Your_request_otp"),
		M360Template:    i18n.T("m360_otp_content"),
	}
	res, err := smsManager.Send(service.CountryCode, service.Mobile, otp)
	if err != nil {
		util.Log().Error("send sms err", err)
	}
	if !res.HasSucceeded {
		return err
	}

	event := model.OtpEvent{
		OtpEventC: models.OtpEventC{
			CountryCode: service.CountryCode,
			Mobile:      service.Mobile,
			Otp:         otp,
			Provider:    utilities.SmsProviderName[res.Provider],
			DateTime:    time.Now().Format(time.DateTime),
		},
	}
	if err := model.LogOtpEvent(event); err != nil {
		// Just log error
		util.Log().Error("log otp event err", err)
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
