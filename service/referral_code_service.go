package service

import (
	"github.com/gin-gonic/gin"
	"strings"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type ReferralCodeVerificationService struct {
	Code string `form:"code" json:"code" binding:"required"`
}

func (service *ReferralCodeVerificationService) Verify(c *gin.Context) serializer.Response {
	service.Code = strings.ToUpper(service.Code)
	i18n := c.MustGet("i18n").(i18n.I18n)
	var existing model.User
	if r := model.DB.Where(`referral_code`, service.Code).Limit(1).Find(&existing).RowsAffected; r == 0 {
		return serializer.ParamErr(c, service, i18n.T("referral_not_found"), nil)
	}
	return serializer.Response{}
}
