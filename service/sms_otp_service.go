package service

import (
	"blgit.rfdev.tech/zhibo/utilities"
	"context"
	"github.com/gin-gonic/gin"
	"math/rand"
	"os"
	"strconv"
	"time"
	"web-api/cache"
	"web-api/serializer"
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
		var provider int
		if os.Getenv("USE_AWS_SNS") == "true" {
			provider = utilities.SMS_PROVIDER_AWSSNS
		} else if os.Getenv("USE_BULKSMS") == "true" {
			provider = utilities.SMS_PROVIDER_BULKSMS
		} else {
			provider = utilities.SMS_PROVIDER_TENCENT
		}
		smsProvider := utilities.SmsProvider{
			Provider:        provider,
			HuanXunTemplate: `您的验证码是 %s，5分钟有效，请尽快验证`,
			BulkSmsTemplate: i18n.T("Your_request_otp"),
			AwsSnsTemplate:  i18n.T("Your_request_otp"),
		}
		if err := smsProvider.Send(service.CountryCode, service.Mobile, otp); err != nil {
			return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("Sms_fail"), err)
		}
	}
	cache.RedisSessionClient.Set(context.TODO(), "otp:"+service.CountryCode+service.Mobile, otp, 2*time.Minute)

	return serializer.Response{
		Msg: i18n.T("success"),
	}
}
