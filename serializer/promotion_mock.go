package serializer

import (
	"time"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

var PromotionMock = []PromotionCover{
	{
		ID:                     1,
		Name:                   "First D B",
		Image:                  "xxx",
		StartAt:                time.Now().Unix(),
		EndAt:                  time.Now().Add(24 * time.Hour).Unix(),
		Type:                   1,
		RewardType:             1,
		RewardDistributionType: 1,
	},
	{
		ID:                     1,
		Name:                   "First D B",
		Image:                  "xxx",
		StartAt:                time.Now().Add(-24 * time.Hour).Unix(),
		EndAt:                  time.Now().Unix(),
		Type:                   1,
		RewardType:             1,
		RewardDistributionType: 1,
	},
}

var PromotionDetailMock = PromotionDetail{
	ID:                     1,
	Name:                   "First D B",
	Image:                  "xxx",
	StartAt:                time.Now().Add(-24 * time.Hour).Unix(),
	EndAt:                  time.Now().Add(24 * time.Hour).Unix(),
	ResetAt:                time.Now().Add(12 * time.Hour).Unix(),
	Type:                   1,
	RewardType:             1,
	RewardDistributionType: 1,
	Reward:                 100,
	PromotionProgress: PromotionProgress{
		Progress: 200,
		Tiers: []RewardTier{
			{
				Min:    0,
				Max:    100,
				Type:   string(models.Fixed),
				Reward: 100,
			},
			{
				Min:    100,
				Max:    200,
				Type:   string(models.Fixed),
				Reward: 200,
			},
			{
				Min:    200,
				Max:    -1,
				Type:   string(models.Fixed),
				Reward: 300,
			},
		},
	},
}
