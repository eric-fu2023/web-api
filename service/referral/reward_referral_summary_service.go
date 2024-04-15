package referral

import (
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

type RewardReferralSummaryService struct {
	ReferralId int64 `form:"referral_id" json:"referral_id" binding:"required"`
}

func (service *RewardReferralSummaryService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	urCond := model.UserReferralCond{
		ReferrerIds: []int64{user.ID},
		ReferralIds: []int64{service.ReferralId},
	}
	urs, err := model.GetUserReferrals(urCond)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetUserReferrals error: %s", err.Error())
		return serializer.GeneralErr(c, err), err
	}
	if len(urs) == 0 {
		return serializer.ParamErr(c, service, i18n.T("referral_not_found"), err), nil
	}

	var referral model.User
	err = model.DB.Where(`id`, service.ReferralId).First(&referral).Error
	if err != nil {
		util.GetLoggerEntry(c).Errorf("get referral user error: %s", err.Error())
		return serializer.GeneralErr(c, err), err
	}

	vip, err := model.GetVipWithDefault(c, service.ReferralId)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetVipWithDefault error: %s", err.Error())
		return serializer.GeneralErr(c, err), err
	}

	cond := model.GetReferralAllianceSummaryCond{ReferralIds: []int64{service.ReferralId}}
	summaries, err := model.GetReferralAllianceSummary(cond)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetReferralAllianceSummary error: %s", err.Error())
		return serializer.GeneralErr(c, err), err
	}
	summary := model.ReferralAllianceSummary{}
	if len(summaries) > 0 {
		summary = summaries[0]
	}

	return serializer.Response{
		Data: serializer.BuildReferralAllianceReferralSummary(referral, vip, summary),
		Msg:  i18n.T("success"),
	}, nil
}
