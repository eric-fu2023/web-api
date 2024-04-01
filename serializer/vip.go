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
	ID                   int64 `json:"id"`
	VIPLevel             int64 `json:"vip_level"`
	WithdrawCount        int64 `json:"withdraw_count"`
	WithdrawAmount       int64 `json:"withdraw_amount"`
	WithdrawAmountTotal  int64 `json:"withdraw_amount_total"`
	PromotionRequirement int64 `json:"promotion_requirement"`
	RetentionRequirement int64 `json:"retention_requirement"`
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
		TotalProgress:   v.TotalProgress,
		CurrentProgress: v.CurrentProgress,
	}

}

func BuildVipRule(v models.VIPRule) VipRule {
	return VipRule{
		ID:                   v.ID,
		VIPLevel:             v.VIPLevel,
		WithdrawCount:        v.WithdrawCount,
		WithdrawAmount:       v.WithdrawAmount,
		WithdrawAmountTotal:  v.WithdrawAmountTotal,
		PromotionRequirement: v.PromotionRequirement,
		RetentionRequirement: v.RetentionRequirement,
	}
}
