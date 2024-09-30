package referral_alliance

import (
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"

	"github.com/gin-gonic/gin"
)

type SummaryService struct{}

func (service *SummaryService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	referralCount, err := model.GetReferralCount(user.ID)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetReferralCount error: %s", err.Error())
		return serializer.GeneralErr(c, err), err
	}

	cond := model.GetReferralAllianceSummaryCond{ReferrerIds: []int64{user.ID}, HasBeenClaimed: []bool{true}}
	summaries, err := model.GetReferralAllianceSummaries(cond)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetReferralAllianceSummaries error: %s", err.Error())
		return serializer.GeneralErr(c, err), err
	}

	summary := model.ReferralAllianceSummary{}
	if len(summaries) > 0 {
		summary = summaries[0]
	}
	claimable_reward := util.Max(summary.ClaimableReward, 0)

	type Response struct {
		ReferralCount     int64   `json:"referral_count"`
		RewardRecordCount int64   `json:"reward_record_count"`
		TotalReward       float64 `json:"total_reward"`
	}

	respData := Response{
		ReferralCount:     referralCount,
		RewardRecordCount: summary.RecordCount,
		TotalReward:       float64(claimable_reward) / 100,
	}

	return serializer.Response{
		Data: respData,
		Msg:  i18n.T("success"),
	}, nil
}
