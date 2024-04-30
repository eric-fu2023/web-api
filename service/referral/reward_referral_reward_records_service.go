package referral

import (
	"github.com/gin-gonic/gin"
	"os"
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
		ReferralIds: []int64{service.ReferralId},
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

	var data map[string]any

	// TODO!Jh remove mocked data for testing
	if os.Getenv("ENV") == "local" || os.Getenv("ENV") == "staging" {
		type RewardRecord struct {
			GameCategoryName string  `json:"game_category_name"`
			ReferrerReward   float64 `json:"referrer_reward"`
		}
		type RewardRecordDay struct {
			Date          string  `json:"date"`
			TotalReward   float64 `json:"total_reward"`
			ClaimedReward float64 `json:"claimed_reward"`
			RewardRecords []RewardRecord
		}

		rewardRecordsDay := []RewardRecordDay{
			{
				Date:          "2024-04-04",
				TotalReward:   308.0,
				ClaimedReward: 108.0,
				RewardRecords: []RewardRecord{
					{
						GameCategoryName: "体育",
						ReferrerReward:   200.0,
					},
					{
						GameCategoryName: "真人",
						ReferrerReward:   58.0,
					},
					{
						GameCategoryName: "电竞",
						ReferrerReward:   50.0,
					},
				},
			},
			{
				Date:          "2024-04-05",
				TotalReward:   245.0,
				ClaimedReward: 125.0,
				RewardRecords: []RewardRecord{
					{
						GameCategoryName: "Slots",
						ReferrerReward:   245.0,
					},
				},
			},
		}

		data = map[string]any{
			"reward_records_day": rewardRecordsDay,
		}
	} else {
		data = map[string]any{
			"reward_records": serializer.BuildReferralAllianceRewards(c, rewardRecords),
		}
	}

	return serializer.Response{
		Data: data,
		Msg:  i18n.T("success"),
	}, nil
}
