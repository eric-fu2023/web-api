package referral

import (
	"github.com/gin-gonic/gin"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

type RewardReferralRewardRecordsService struct {
	ReferralId      int64 `form:"referral_id"  binding:"required"`
	RecordTimeStart int64 `form:"record_time_start"`
	RecordTimeEnd   int64 `form:"record_time_end"`
}

func (service *RewardReferralRewardRecordsService) List(c *gin.Context) (r serializer.Response, err error) {
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

	cond := model.GetReferralAllianceRewardsCond{
		ReferralIds:    []int64{service.ReferralId},
		HasBeenClaimed: []bool{true},
	}
	if service.RecordTimeStart > 0 {
		cond.BetDateStart = time.Unix(service.RecordTimeStart, 0).Format(time.DateOnly)
	}
	if service.RecordTimeEnd > 0 {
		cond.BetDateEnd = time.Unix(service.RecordTimeEnd, 0).Format(time.DateOnly)
	}
	rewardRecords, err := model.GetReferralAllianceRewards(cond)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetReferralAllianceRewards error: %s", err.Error())
		return serializer.GeneralErr(c, err), err
	}

	return serializer.Response{
		Data: map[string]any{
			"reward_records_day": serializer.BuildReferralAllianceRewards(c, rewardRecords),
		},
		Msg: i18n.T("success"),
	}, nil
}
