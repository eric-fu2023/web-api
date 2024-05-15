package referral_alliance

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
)

type ReferralsService struct {
	JoinTimeStart int64 `form:"join_time_start"`
	JoinTimeEnd   int64 `form:"join_time_end"`
}

func (service *ReferralsService) List(c *gin.Context) (r serializer.Response, err error) {
	u, _ := c.Get("user")
	user := u.(model.User)

	rdCond := model.GetReferralDetailsCond{ReferrerId: user.ID}
	if service.JoinTimeStart != 0 {
		rdCond.JoinTimeStart = sql.NullTime{Time: time.Unix(service.JoinTimeStart, 0), Valid: true}
	}
	if service.JoinTimeEnd != 0 {
		rdCond.JoinTimeEnd = sql.NullTime{Time: time.Unix(service.JoinTimeEnd, 0), Valid: true}
	}

	referralDetails, err := model.GetReferralDetails(rdCond)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetReferralDetails error: %s", err.Error())
		return serializer.GeneralErr(c, err), err
	}

	defaultVip, err := model.GetDefaultVip()
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetDefaultVip error: %s", err.Error())
		return serializer.GeneralErr(c, err), err
	}

	var referralIds []int64
	for _, rd := range referralDetails {
		referralIds = append(referralIds, rd.ReferralId)
	}

	rewardSummaries, err := model.GetReferralAllianceSummaries(model.GetReferralAllianceSummaryCond{
		ReferralIds:    referralIds,
		HasBeenClaimed: []bool{true},
	})
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetReferralAllianceSummaries error: %s", err.Error())
		return serializer.GeneralErr(c, err), err
	}

	rewardSummaryMap := map[int64]model.ReferralAllianceSummary{}
	for _, rs := range rewardSummaries {
		rewardSummaryMap[rs.ReferralId] = rs
	}

	return serializer.Response{
		Data: map[string]any{
			"referrals": serializer.BuildReferralAllianceReferrals(c, referralDetails, rewardSummaryMap, defaultVip),
		},
	}, nil
}
