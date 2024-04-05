package referral

import (
	"github.com/gin-gonic/gin"
	"web-api/serializer"
	"web-api/util/i18n"
)

type RewardReferralsService struct {
	JoinTimeStart int64 `form:"join_time_start"`
	JoinTimeEnd   int64 `form:"join_time_end"`
	Limit         int64 `form:"limit"  binding:"required"`
}

func (service *RewardReferralsService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	//u, _ := c.Get("user")
	//user := u.(model.User)
	// TODO!Jh replace mock

	type Referral struct {
		Id             int64   `json:"id"`
		Nickname       string  `json:"nickname"`
		Avatar         string  `json:"avatar"`
		VipId          int64   `json:"vip_id"`
		JoinTime       int64   `json:"join_time"`
		ReferrerReward float64 `json:"referrer_reward"`
	}

	type Response struct {
		Referrals []Referral `json:"referrals"`
	}

	respData := Response{
		Referrals: []Referral{
			{
				Id:             224,
				Nickname:       "Some User",
				Avatar:         "https://static.tayalive.com/img/user/224/avatar/224-avatar-20240404090035-PWauUp.jpg",
				VipId:          1,
				JoinTime:       1711617294,
				ReferrerReward: 996.03,
			},
			{
				Id:             225,
				Nickname:       "Another User",
				VipId:          2,
				JoinTime:       1711617293,
				ReferrerReward: 1080.01,
			},
		},
	}

	return serializer.Response{
		Data: respData,
		Msg:  i18n.T("success"),
	}, nil
}
