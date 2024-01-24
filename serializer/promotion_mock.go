package serializer

import "time"

var PromotionMock = []PromotionCover{
	{
		ID:                     1,
		Name:                   "First D B",
		Description:            "Deposit now",
		Image:                  "xxx",
		StartAt:                time.Now(),
		EndAt:                  time.Now().Add(24 * time.Hour),
		Type:                   1,
		RewardType:             1,
		RewardDistributionType: 1,
	},
	{
		ID:                     1,
		Name:                   "First D B",
		Description:            "Deposit now",
		Image:                  "xxx",
		StartAt:                time.Now().Add(-24 * time.Hour),
		EndAt:                  time.Now(),
		Type:                   1,
		RewardType:             1,
		RewardDistributionType: 1,
	},
}

var PromotionDetailMock = PromotionDetail{
	ID:                     1,
	Name:                   "First D B",
	Description:            "Deposit now",
	Image:                  "xxx",
	StartAt:                time.Now().Add(-24 * time.Hour),
	EndAt:                  time.Now().Add(24 * time.Hour),
	ResetAt:                time.Now().Add(12 * time.Hour),
	Type:                   1,
	RewardType:             1,
	RewardDistributionType: 1,
	Reward:                 100,
	IsEligible:             true,
	PromotionProgress: PromotionProgress{
		Progress: 200,
		Tiers: []PromotionTier{
			{
				Min:    0,
				Max:    100,
				Reward: 100,
			},
			{
				Min:    100,
				Max:    200,
				Reward: 200,
			},
			{
				Min:    200,
				Max:    -1,
				Reward: 300,
			},
		},
	},
}
