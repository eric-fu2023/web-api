package util

import (
	smsutil "blgit.rfdev.tech/zhibo/utilities/sms"
	"github.com/gin-gonic/gin"
	"strconv"
	"web-api/util/i18n"
)

func BuildSmsTemplates(c *gin.Context) smsutil.Templates {
	i18n := c.MustGet("i18n").(i18n.I18n)
	brandId := c.MustGet("_brand").(int)

	if brandId >= 3000 && brandId < 5000 {
		brandId = brandId / 1000 * 1000 // round down to nearest thousand
	}

	return smsutil.Templates{
		Default: i18n.T("Your_request_otp" + "." + strconv.FormatInt(int64(brandId), 10)),
		M360:    i18n.T("m360_otp_content" + "." + strconv.FormatInt(int64(brandId), 10)),
		CxCloud: "%s", // Only need OTP. Msg will be automatically formatted by provider
	}
}
