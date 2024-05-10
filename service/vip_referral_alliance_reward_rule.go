package service

import (
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
)

type VipReferralAllianceRewardRulesService struct{}

func (v VipReferralAllianceRewardRulesService) Load(c *gin.Context) (r serializer.Response, err error) {
	list, err := model.GetAllReferralAllianceRules()
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Err GetAllReferralAllianceRules: %s", err.Error())
		r = serializer.Err(c, "", serializer.CodeGeneralError, "", err)
		return
	}
	vips, err := model.LoadVipRule(c)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("Err LoadVipRule: %s", err.Error())
		r = serializer.Err(c, "", serializer.CodeGeneralError, "", err)
		return
	}
	desc, err := model.GetAppConfig("referral_alliance", "vip_details_description")
	if err != nil {
		r = serializer.Err(c, "", serializer.CodeGeneralError, "", err)
		return
	}
	r.Data = serializer.BuildVipReferralDetails(c, list, desc, vips)
	return
}
