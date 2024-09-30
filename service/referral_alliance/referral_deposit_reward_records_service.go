package referral_alliance

import (
	"fmt"
	"strconv"
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"

	"github.com/gin-gonic/gin"
)

type ReferralDepositRewardRecordsService struct {
	ReferralId      int64 `form:"referral_id"  binding:"required"`
	RecordTimeStart int64 `form:"record_time_start"`
	RecordTimeEnd   int64 `form:"record_time_end"`
}

func (service *ReferralDepositRewardRecordsService) List(c *gin.Context) (r serializer.Response, err error) {
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
		monthStartStr, err := service.getDateString(time.Unix(service.RecordTimeStart, 0))
		if err != nil {
			util.GetLoggerEntry(c).Errorf("getMonthString start error: %s", err.Error())
			return serializer.GeneralErr(c, err), err
		}
		cond.RewardMonthStart = monthStartStr
	}
	if service.RecordTimeEnd > 0 {
		monthEndStr, err := service.getDateString(time.Unix(service.RecordTimeEnd, 0))
		if err != nil {
			util.GetLoggerEntry(c).Errorf("getMonthString end error: %s", err.Error())
			return serializer.GeneralErr(c, err), err
		}
		cond.RewardMonthEnd = monthEndStr
	}

	rewardRecords, err := model.GetReferralAllianceRewards(cond)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetReferralAllianceRewards error: %s", err.Error())
		return serializer.GeneralErr(c, err), err
	}

	return serializer.Response{
		Data: map[string]any{
			"reward_records_month": serializer.BuildReferralDepositAllianceRewards(c, rewardRecords),
		},
		Msg: i18n.T("success"),
	}, nil
}

func (service *ReferralDepositRewardRecordsService) getDateString(t time.Time) (string, error) {
	tzOffsetStr, err := model.GetAppConfigWithCache("timezone", "offset_seconds")
	if err != nil {
		return "", fmt.Errorf("failed to get tz offset config: %w", err)
	}
	tzOffset, err := strconv.Atoi(tzOffsetStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse tz offset config: %w", err)
	}

	return t.In(time.FixedZone("", tzOffset)).Format(consts.StdDateFormat), nil
}
