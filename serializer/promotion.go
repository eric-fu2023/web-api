package serializer

import (
	"encoding/json"
	"strconv"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

type PromotionCover struct {
	ID                     int64           `json:"id"`
	Name                   string          `json:"name"`
	Description            json.RawMessage `json:"description"`
	Image                  string          `json:"image"`
	StartAt                int64           `json:"start_at"`
	EndAt                  int64           `json:"end_at"`
	Type                   int64           `json:"type"`
	RewardType             int64           `json:"reward_type"`
	RewardDistributionType int64           `json:"reward_distribution_type"`
}

type PromotionDetail struct {
	ID                     int64             `json:"id"`
	Name                   string            `json:"name"`
	Description            json.RawMessage   `json:"description"`
	Image                  string            `json:"image"`
	StartAt                int64             `json:"start_at"`
	EndAt                  int64             `json:"end_at"`
	ResetAt                int64             `json:"reset_at"`
	Type                   int64             `json:"type"`
	RewardType             int64             `json:"reward_type"`
	RewardDistributionType int64             `json:"reward_distribution_type"`
	PromotionProgress      PromotionProgress `json:"promotion_progress"`
	Reward                 float64           `json:"reward"`
	ClaimStatus            ClaimStatus       `json:"claim_status"`
	Voucher                Voucher           `json:"voucher"`
}

type ClaimStatus struct {
	HasClaimed bool  `json:"has_claimed"`
	ClaimStart int64 `json:"claim_start"`
	ClaimEnd   int64 `json:"claim_end"`
}

type PromotionProgress struct {
	Progress float64      `json:"progress"`
	Tiers    []RewardTier `json:"tiers"`
}

type RewardTier struct {
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Type   string  `json:"type"`
	Reward float64 `json:"reward"`
}

func BuildPromotionCover(p models.Promotion, platform string) PromotionCover {
	raw := json.RawMessage(p.Image)
	m := make(map[string]string)
	json.Unmarshal(raw, &m)
	image := m[platform]
	if len(image) == 0 {
		image = m["h5"]
	}
	return PromotionCover{
		ID:                     p.ID,
		Name:                   p.Name,
		Description:            json.RawMessage(p.Description),
		Image:                  Url(image),
		StartAt:                p.StartAt.Unix(),
		EndAt:                  p.EndAt.Unix(),
		Type:                   p.Type,
		RewardType:             p.RewardType,
		RewardDistributionType: p.RewardDistributionType,
	}
}

func BuildPromotionDetail(progress, reward int64, platform string, p models.Promotion, s models.PromotionSession, v Voucher) PromotionDetail {
	raw := json.RawMessage(p.Image)
	m := make(map[string]string)
	json.Unmarshal(raw, &m)
	image := m[platform]
	if len(image) == 0 {
		image = m["h5"]
	}
	return PromotionDetail{
		ID:                     p.ID,
		Name:                   p.Name,
		Description:            json.RawMessage(p.Description),
		Image:                  Url(image),
		StartAt:                p.StartAt.Unix(),
		EndAt:                  p.EndAt.Unix(),
		ResetAt:                s.EndAt.Unix(),
		Type:                   p.Type,
		RewardType:             p.RewardType,
		RewardDistributionType: p.RewardDistributionType,
		PromotionProgress:      BuildPromotionProgress(progress, p.GetRewardDetails()),
		Reward:                 float64(reward) / 100,
		Voucher:                v,
	}
}

func BuildPromotionProgress(progress int64, rewards models.RewardDetails) PromotionProgress {
	return PromotionProgress{
		Progress: float64(progress) / 100,
		Tiers:    util.MapSlice(rewards.Rewards, buildPromotionTier),
	}
}

func buildPromotionTier(rewardTier models.TierdReward) RewardTier {
	var (
		min float64
		max float64
	)
	for _, c := range rewardTier.Conditions {
		if c.Operator == models.Gt || c.Operator == models.Gte {
			min, _ = strconv.ParseFloat(c.Value, 64)
		} else if c.Operator == models.Lt || c.Operator == models.Lte {
			max, _ = strconv.ParseFloat(c.Value, 64)
		}
	}
	p := RewardTier{
		Min:    min / 100,
		Max:    max / 100,
		Type:   string(rewardTier.Rewards[0].Type),
		Reward: float64(rewardTier.Rewards[0].Value) / 100,
	}
	return p
}
