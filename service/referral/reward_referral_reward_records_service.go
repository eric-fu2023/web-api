package referral

import (
	"github.com/gin-gonic/gin"
	"web-api/serializer"
	"web-api/util/i18n"
)

type RewardReferralRewardRecordsService struct {
	ReferralId      int64 `form:"referral_id"  binding:"required"`
	RecordTimeStart int64 `form:"record_time_start"`
	RecordTimeEnd   int64 `form:"record_time_end"`
}

func (service *RewardReferralRewardRecordsService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	//u, _ := c.Get("user")
	//user := u.(model.User)
	// TODO!Jh replace mock
	type RewardRecord struct {
		Date             string  `json:"date"`
		GameCategoryName string  `json:"game_category_name"`
		ReferrerReward   float64 `json:"referrer_reward"`
	}

	type Response struct {
		RewardRecords []RewardRecord `json:"reward_records"`
	}

	respData := Response{
		RewardRecords: []RewardRecord{
			{
				Date:             "2024-04-04",
				GameCategoryName: "Sports",
				ReferrerReward:   888.01,
			},
			{
				Date:             "2024-04-04",
				GameCategoryName: "Slots",
				ReferrerReward:   108.02,
			},
		},
	}

	return serializer.Response{
		Data: respData,
		Msg:  i18n.T("success"),
	}, nil
}
