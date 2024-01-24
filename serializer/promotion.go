package serializer

import "time"

type PromotionCover struct {
	ID                     int64     `json:"id"`
	Name                   string    `json:"name"`
	Description            string    `json:"description"`
	Image                  string    `json:"image"`
	StartAt                time.Time `json:"start_at"`
	EndAt                  time.Time `json:"end_at"`
	Type                   int64     `json:"type"`
	RewardType             int64     `json:"reward_type"`
	RewardDistributionType int64     `json:"reward_distribution_type"`
}

type PromotionDetail struct {
	ID                     int64             `json:"id"`
	Name                   string            `json:"name"`
	Description            string            `json:"description"`
	Image                  string            `json:"image"`
	StartAt                time.Time         `json:"start_at"`
	EndAt                  time.Time         `json:"end_at"`
	ResetAt                time.Time         `json:"reset_at"`
	Type                   int64             `json:"type"`
	RewardType             int64             `json:"reward_type"`
	RewardDistributionType int64             `json:"reward_distribution_type"`
	PromotionProgress      PromotionProgress `json:"promotion_progress"`
	Reward                 int64             `json:"reward"`
	IsEligible             bool              `json:"is_eligible"`
}

type PromotionProgress struct {
	Progress int64           `json:"progress"`
	Tiers    []PromotionTier `json:"tiers"`
}

type PromotionTier struct {
	Min    int64 `json:"min,omitempty"`
	Max    int64 `json:"max"`
	Reward int64 `json:"reward"`
}
