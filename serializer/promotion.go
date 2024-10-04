package serializer

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"web-api/model"

	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

type OutgoingEarnMoreMissionTier struct {
	MissionId     int64  `json:"mission_id"`
	MissionAmount int64  `json:"mission_amount"`
	Label         string `json:"label"`
	RewardAmount  int64  `json:"reward_amount"`
	Status        int64  `json:"status"` // go = 0, claim = 1, complete = 2
	CurrentAmount int64  `json:"current_amount"`
}

type OutgoingEarnMoreCardDetail struct {
	Title string `json:"title"`
	Icon  string `json:"icon_url"`
}

type OutgoingEarnMoreMission struct {
	Name                      string                        `json:"name"`
	Desc                      string                        `json:"desc"`
	BackgroundImgUrl          string                        `json:"bg_url"`
	Tooltip                   string                        `json:"tooltip_text"`
	TotalDepositAmount        int64                         `json:"deposit_amount"`
	Card                      OutgoingEarnMoreCardDetail    `json:"card"`
	Label                     string                        `json:"label"`
	DepositStartDate          int64                         `json:"deposit_start_ts"`
	DepositEndDate            int64                         `json:"deposit_end_ts"`
	PromotionDisplayStartDate int64                         `json:"promotion_start_ts"`
	PromotionDisplayEndDate   int64                         `json:"promotion_end_ts"`
	Missions                  []OutgoingEarnMoreMissionTier `json:"missions"`
}

type PromotionCover struct {
	ID                     int64           `json:"id"`
	Name                   string          `json:"name"`
	Title                  string          `json:"title"`
	Description            json.RawMessage `json:"description"`
	Image                  string          `json:"image"`
	StartAt                int64           `json:"start_at"`
	EndAt                  int64           `json:"end_at"`
	Type                   int64           `json:"type"`
	RewardType             int64           `json:"reward_type"`
	RewardDistributionType int64           `json:"reward_distribution_type"`
	Category               int64           `json:"category"`
	Label                  int64           `json:"label"`
	IsVipAssociated        bool            `json:"is_vip_associated"`
	DisplayOnly            bool            `json:"display_only"`
	ParentId               int64           `json:"parent_id"`
	IsCustom               bool            `json:"is_custom"`
	SubpageContent         json.RawMessage `json:"subpage_content"`

	ChildrenPromotions []PromotionCover `json:"children_promotions"`
}

type PromotionDetail struct {
	ID                     int64                   `json:"id"`
	Name                   string                  `json:"name"`
	Description            json.RawMessage         `json:"description"`
	Image                  string                  `json:"image"`
	StartAt                int64                   `json:"start_at"`
	EndAt                  int64                   `json:"end_at"`
	RecurringDay           int64                   `json:"recurring_day"`
	ResetAt                int64                   `json:"reset_at"`
	Type                   int64                   `json:"type"`
	RewardType             int64                   `json:"reward_type"`
	RewardDistributionType int64                   `json:"reward_distribution_type"`
	Category               int64                   `json:"category"`
	Label                  int64                   `json:"label"`
	PromotionProgress      PromotionProgress       `json:"promotion_progress"`
	Reward                 float64                 `json:"reward"`
	ClaimStatus            ClaimStatus             `json:"claim_status"`
	Voucher                Voucher                 `json:"voucher"`
	IsVipAssociated        bool                    `json:"is_vip_associated"`
	DisplayOnly            bool                    `json:"display_only"`
	Extra                  any                     `json:"extra"`
	CustomTemplateData     json.RawMessage         `json:"custom_template_data"`
	NewbieData             interface{}             `json:"newbie_data"` // TODO : to be updated
	EarnMoreData           OutgoingEarnMoreMission `json:"earn_more_promotion_data"`

	IsCustom bool `json:"is_custom"`
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

type SysDictionaryDetail struct {
	ID    int64  `json:"id"`
	Label string `json:"label" form:"label" gorm:"column:label;comment:展示值"` // 展示值
	Value int    `json:"value" form:"value" gorm:"column:value;comment:字典值"` // 字典值
	Sort  int    `json:"sort" form:"sort" gorm:"column:sort;comment:排序标记"`   // 排序标记
}

type MissionDO struct {
	Missions           []model.MissionTier `json:"missions"`
	CompletedMissions  []models.Voucher    `json:"completed_missions"`
	TotalDepositAmount int64               `json:"total_deposit_amount"`
}

func BuildSysDictionaryDetail(s models.SysDictionaryDetail) SysDictionaryDetail {
	return SysDictionaryDetail{
		ID:    s.ID,
		Label: s.Label,
		Value: s.Value,
		Sort:  s.Sort,
	}
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
		Description:            p.Description,
		Image:                  Url(image),
		StartAt:                p.StartAt.Unix(),
		EndAt:                  p.EndAt.Unix(),
		Type:                   p.Type,
		RewardType:             p.RewardType,
		RewardDistributionType: int64(p.RewardDistributionType),
		Category:               int64(p.Category),
		Label:                  int64(p.Label),
		IsVipAssociated:        p.VipAssociated,
		DisplayOnly:            p.DisplayOnly,
		ParentId:               p.ParentId,

		SubpageContent: p.SubpageContent,
	}
}

func BuildPromotionDetail(progress, reward int64, platform string, p models.Promotion, s models.PromotionSession, voucher Voucher, cl ClaimStatus, extra any, customData any, newbieData any, mDo MissionDO) PromotionDetail {
	raw := json.RawMessage(p.Image)
	m := make(map[string]string)
	json.Unmarshal(raw, &m)
	image := m[platform]
	if len(image) == 0 {
		image = m["h5"]
	}

	var earnMoreData OutgoingEarnMoreMission

	if len(mDo.Missions) > 0 {

		// totalDepositedAmount := mDo.TotalDepositAmount / 100
		totalDepositedAmount := int64(rand.Intn(1000000)) / 100

		card := OutgoingEarnMoreCardDetail{
			Title: "Deposit Insights",
			Icon:  "https://static.tayalive.com/batace-img/icon/Evolution.png",
		}

		var earnMoreMissionTiers []OutgoingEarnMoreMissionTier

		for i, mission := range mDo.Missions {
			m := OutgoingEarnMoreMissionTier{
				MissionId:     int64(i),
				MissionAmount: mission.MissionAmount / 100,
				RewardAmount:  mission.RewardAmount / 100,
				Label:         "Deposit ₹" + fmt.Sprint(int(mission.MissionAmount/100)),
				Status:        models.PromotionMissionPendingStatus,
				CurrentAmount: int64(math.Min(float64(totalDepositedAmount), float64(mission.MissionAmount/100))),
			}

			if m.CurrentAmount == m.MissionAmount {
				m.Status = models.PromotionMissionReadyStatus
			}

			for _, completed := range mDo.CompletedMissions {
				if completed.ReferenceID == fmt.Sprint(m.MissionId) {
					m.Status = models.PromotionMissionCompletedStatus
					break
				}
			}

			earnMoreMissionTiers = append(earnMoreMissionTiers, m)
		}

		earnMoreData = OutgoingEarnMoreMission{
			Name:                      p.Name,
			Tooltip:                   "TOOLTIP",
			TotalDepositAmount:        totalDepositedAmount,
			Label:                     "Deposit",
			DepositStartDate:          1727782019,
			DepositEndDate:            1730287619,
			PromotionDisplayStartDate: 1727782019,
			PromotionDisplayEndDate:   1730287619,
			Card:                      card,
			Desc:                      "Boost your deposit, unlock bigger rewards!",
			Missions:                  earnMoreMissionTiers,
		}
	}

	return PromotionDetail{
		ID:                     p.ID,
		Name:                   p.Name,
		Description:            p.Description,
		Image:                  Url(image),
		StartAt:                p.StartAt.Unix(),
		EndAt:                  p.EndAt.Unix(),
		ResetAt:                s.EndAt.Unix(),
		Type:                   p.Type,
		RewardType:             p.RewardType,
		RecurringDay:           int64(p.RecurringDay),
		RewardDistributionType: int64(p.RewardDistributionType),
		ClaimStatus:            cl,
		PromotionProgress:      BuildPromotionProgress(progress, p.GetRewardDetails()),
		Reward:                 float64(reward) / 100,
		Voucher:                voucher,
		Category:               int64(p.Category),
		IsVipAssociated:        p.VipAssociated,
		DisplayOnly:            p.DisplayOnly,
		Extra:                  extra,
		// CustomTemplateData: 	json.RawMessage(customData),
		NewbieData:   newbieData,
		EarnMoreData: earnMoreData,
	}
}

func BuildPromotionProgress(progress int64, tieredRewards models.RewardDetails) PromotionProgress {
	return PromotionProgress{
		Progress: float64(progress) / 100,
		Tiers:    util.MapSlice(tieredRewards.Rewards, buildPromotionTier),
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

// func BuildJoinCustomPromotionRequest(p PromotionJoin, request models.PromotionRequest) {

// }
