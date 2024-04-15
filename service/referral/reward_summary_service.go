package referral

import (
	"github.com/gin-gonic/gin"
	"web-api/serializer"
	"web-api/util/i18n"
)

type RewardSummaryService struct{}

func (service *RewardSummaryService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	//u, _ := c.Get("user")
	//user := u.(model.User)
	// TODO!Jh replace mock

	type Response struct {
		ReferralCount     int64   `json:"referral_count"`
		RewardRecordCount int64   `json:"reward_record_count"`
		TotalReward       float64 `json:"total_reward"`
	}

	respData := Response{
		ReferralCount:     2,
		RewardRecordCount: 5,
		TotalReward:       2076.04,
	}

	return serializer.Response{
		Data: respData,
		Msg:  i18n.T("success"),
	}, nil
}
