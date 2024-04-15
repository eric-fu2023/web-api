package referral

import (
	"github.com/gin-gonic/gin"
	"web-api/serializer"
	"web-api/util/i18n"
)

type RewardReferralSummaryService struct {
	ReferralId int64 `form:"referral_id" json:"referral_id" binding:"required"`
}

func (service *RewardReferralSummaryService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	//u, _ := c.Get("user")
	//user := u.(model.User)
	// TODO!Jh replace mock

	type Response struct {
		ReferralId        int64   `json:"referral_id"`
		Nickname          string  `json:"nickname"`
		Avatar            string  `json:"avatar"`
		VipId             int64   `json:"vip_id"`
		JoinTime          int64   `json:"join_time"`
		RewardRecordCount int64   `json:"reward_record_count"`
		TotalReward       float64 `json:"total_reward"`
	}

	respData := Response{
		ReferralId:        224,
		Nickname:          "Some User",
		Avatar:            "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
		VipId:             3,
		JoinTime:          1711617294,
		RewardRecordCount: 2,
		TotalReward:       996.03,
	}

	return serializer.Response{
		Data: respData,
		Msg:  i18n.T("success"),
	}, nil
}
