package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/util"
)

type ReferralAllianceReferralSummary struct {
	ReferralId  int64   `json:"referral_id"`
	Nickname    string  `json:"nickname"`
	Avatar      string  `json:"avatar"`
	VipId       int64   `json:"vip_id"`
	JoinTime    int64   `json:"join_time"`
	TotalReward float64 `json:"total_reward"`
}

func BuildReferralAllianceReferralSummary(referral model.User, vip ploutos.VipRecord, rewardSummary model.ReferralAllianceSummary) ReferralAllianceReferralSummary {
	return ReferralAllianceReferralSummary{
		ReferralId:  referral.ID,
		Nickname:    referral.Nickname,
		Avatar:      Url(referral.Avatar),
		VipId:       vip.VipRule.ID,
		JoinTime:    referral.CreatedAt.Unix(),
		TotalReward: float64(rewardSummary.TotalReward) / 100,
	}
}

type ReferralAllianceReward struct {
	GameCategoryName string  `json:"game_category_name"`
	ReferrerReward   float64 `json:"referrer_reward"`
}

type ReferralAllianceRewardMonth struct {
	Month         string                   `json:"month"`
	TotalReward   float64                  `json:"total_reward"`
	ClaimedReward float64                  `json:"claimed_reward"`
	RewardRecords []ReferralAllianceReward `json:"reward_records"`
}

func BuildReferralAllianceRewards(c *gin.Context, rewardRecords []ploutos.ReferralAllianceReward) []ReferralAllianceRewardMonth {
	// group by month
	var monthMap = make(map[string][]ploutos.ReferralAllianceReward)
	for _, r := range rewardRecords {
		monthMap[r.RewardMonth] = append(monthMap[r.RewardMonth], r)
	}

	var rewardsMonth []ReferralAllianceRewardMonth
	for rewardMonth, dbRewards := range monthMap {
		var claimableSum int64 = 0
		var totalRewardSum int64 = 0
		var resRewards []ReferralAllianceReward
		for _, r := range dbRewards {
			claimableSum += r.ClaimableAmount
			totalRewardSum += r.Amount
			resRewards = append(resRewards, ReferralAllianceReward{
				GameCategoryName: FormatGameCategoryName(c, r.GameCategoryID),
				ReferrerReward:   float64(r.Amount) / 100,
			})
		}

		rewardsMonth = append(rewardsMonth, ReferralAllianceRewardMonth{
			Month:         rewardMonth,
			TotalReward:   float64(totalRewardSum) / 100,
			ClaimedReward: float64(claimableSum) / 100,
			RewardRecords: resRewards,
		})
	}

	return rewardsMonth
}

type ReferralAllianceReferral struct {
	Id             int64   `json:"id"`
	Nickname       string  `json:"nickname"`
	Avatar         string  `json:"avatar"`
	VipId          int64   `json:"vip_id"`
	JoinTime       int64   `json:"join_time"`
	ReferrerReward float64 `json:"referrer_reward"`
}

func BuildReferralAllianceReferrals(
	c *gin.Context,
	referralDetails []model.UserReferral,
	rewardSummaries map[int64]model.ReferralAllianceSummary,
	defaultVip ploutos.VIPRule,
) []ReferralAllianceReferral {
	var resp []ReferralAllianceReferral
	for _, rd := range referralDetails {
		rs := model.ReferralAllianceSummary{}
		if v, ok := rewardSummaries[rd.ReferralId]; ok {
			rs = v
		}
		if rd.ReferralVipRecord == nil {
			rd.ReferralVipRecord = &ploutos.VipRecord{
				VipRule: defaultVip,
			}
		}

		if rd.Referral == nil {
			util.GetLoggerEntry(c).Errorf("BuildReferralAllianceReferrals: Referral should not be nil: %d", rd.ReferralId)
			continue
		}

		resp = append(resp, ReferralAllianceReferral{
			Id:             rd.Referral.ID,
			Nickname:       rd.Referral.Nickname,
			Avatar:         Url(rd.Referral.Avatar),
			VipId:          rd.ReferralVipRecord.VipRule.ID,
			JoinTime:       rd.Referral.CreatedAt.Unix(),
			ReferrerReward: float64(rs.TotalReward) / 100,
		})
	}
	return resp
}
