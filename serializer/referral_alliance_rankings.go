package serializer

type ReferralAllianceRankings struct {
	RewardRankings   []ReferralAllianceRanking `json:"reward_rankings,omitempty"`
	ReferralRankings []ReferralAllianceRanking `json:"referral_rankings,omitempty"`
	LastUpdateTime   int64                     `json:"last_update_time,omitempty"`
}

type ReferralAllianceRanking struct {
	ReferralAllianceRankingUserDetails
	Rank int64 `json:"rank"`
}

type ReferralAllianceRankingUserRanking struct {
	ReferralAllianceRankingUserDetails
	RewardRank   int64 `json:"reward_rank"`
	ReferralRank int64 `json:"referral_rank"`
}

type ReferralAllianceRankingUserDetails struct {
	Id            int64   `json:"id"`
	Nickname      string  `json:"nickname"`
	Avatar        string  `json:"avatar"`
	RewardAmount  float64 `json:"reward_amount,omitempty"`
	ReferralCount int64   `json:"referral_count,omitempty"`
}
