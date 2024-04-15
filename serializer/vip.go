package serializer

import (
	"time"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

type Vip struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	AcquiredAt time.Time `json:"acquired_at"`
	ExpireAt   time.Time `json:"expire_at"`

	Progress VipProgress `json:"progress"`
	Rule     VipRule     `json:"rule"`
}

type VipProgress struct {
	ID              int64 `json:"id"`
	UserID          int64 `json:"user_id"`
	TotalProgress   int64 `json:"total_progress"`
	CurrentProgress int64 `json:"current_progress"`
}

type VipRule struct {
	ID                  int64 `json:"id"`
	VIPLevel            int64 `json:"vip_level"`
	WithdrawCount       int64 `json:"withdraw_count"`
	WithdrawAmount      int64 `json:"withdraw_amount"`
	WithdrawAmountTotal int64 `json:"withdraw_amount_total"`
	// PromotionRequirement int64   `json:"promotion_requirement"`
	TotalRequirement     int64   `json:"total_requirement"`
	RetentionRequirement int64   `json:"retention_requirement"`
	Icon                 string  `json:"icon"`
	Background           string  `json:"background"`
	RebateRate           float64 `json:"rebate_rate"`
	RebateCap            int64   `json:"rebate_cap"`
	ReferralRate         float64 `json:"referral_rate"`
	ReferralCap          int64   `json:"referral_cap"`
	PromotionBenefit     int64   `json:"promotion_benefit"`
	WeeklyBenefit        int64   `json:"weekly_benefit"`
	BirthdayBenefit      int64   `json:"birthday_benefit"`
}

func BuildVip(v models.VipRecord) Vip {
	return Vip{
		ID:         v.ID,
		UserID:     v.UserID,
		AcquiredAt: v.AcquiredAt,
		ExpireAt:   v.ExpireAt,
		Progress:   BuildVipProgress(v.VipProgress),
		Rule:       BuildVipRule(v.VipRule),
	}
}

func BuildVipProgress(v models.VipProgress) VipProgress {
	return VipProgress{
		ID:              v.ID,
		UserID:          v.UserID,
		TotalProgress:   v.TotalProgress / 100,
		CurrentProgress: v.CurrentProgress / 100,
	}

}

func BuildVipRule(v models.VIPRule) VipRule {
	return VipRule{
		ID:                   v.ID,
		VIPLevel:             v.VIPLevel,
		WithdrawCount:        v.WithdrawCount,
		WithdrawAmount:       v.WithdrawAmount / 100,
		WithdrawAmountTotal:  v.WithdrawAmountTotal / 100,
		TotalRequirement:     v.TotalRequirement / 100,
		RetentionRequirement: v.RetentionRequirement / 100,
		Icon:                 Url(v.Icon),
		Background:           v.Background,
		RebateRate:           v.RebateRate,
		ReferralRate:         v.ReferralRate,
		PromotionBenefit:     v.PromotionBenefit / 100,
		WeeklyBenefit:        v.WeeklyBenefit / 100,
		BirthdayBenefit:      v.BirthdayBenefit / 100,
		RebateCap:            v.RebateCap / 100,
		ReferralCap:          v.ReferralCap / 100,
	}
}
