package service

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"blgit.rfdev.tech/zhibo/utilities"
	"context"
	"github.com/gin-gonic/gin"
	"math/rand"
	"os"
	"strconv"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
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

	var provider int
	if os.Getenv("USE_AWS_SNS") == "true" {
		provider = utilities.SMS_PROVIDER_AWSSNS
	} else if os.Getenv("USE_BULKSMS") == "true" {
		provider = utilities.SMS_PROVIDER_BULKSMS
	} else if os.Getenv("USE_TENCENT_SMS") == "true" {
		provider = utilities.SMS_PROVIDER_TENCENT
	}

	if provider != 0 {
		smsProvider := utilities.SmsProvider{
			Provider:        provider,
			HuanXunTemplate: `您的验证码是 %s，5分钟有效，请尽快验证`,
			BulkSmsTemplate: i18n.T("Your_request_otp"),
			AwsSnsTemplate:  i18n.T("Your_request_otp"),
		}
		if err := smsProvider.Send(service.CountryCode, service.Mobile, otp); err != nil {
			return err
		}

		event := model.OtpEvent{
			OtpEventC: models.OtpEventC{
				CountryCode: service.CountryCode,
				Mobile:      service.Mobile,
				Otp:         otp,
				Provider:    utilities.SmsProviderName[provider],
				DateTime:    time.Now().Format(time.DateTime),
			},
		}
		if err := model.LogOtpEvent(event); err != nil {
			// Just log error
			util.Log().Error("log otp event err", err)
		}
	}

	return nil
}
