package serializer

import (
	"time"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

type Vip struct {
	ID         int64
	UserID     int64
	AcquiredAt time.Time
	ExpireAt   time.Time

	Progress VipProgress
	Rule     VipRule
}

type VipProgress struct {
	ID              int64
	UserID          int64
	TotalProgress   int64
	CurrentProgress int64
}

type VipRule struct {
	ID                   int64 `json:"id"`
	VIPLevel             int64 `json:"vip_level" gorm:"default:0"`
	WithdrawCount        int64
	WithdrawAmount       int64
	WithdrawAmountTotal  int64
	PromotionRequirement int64
	RetentionRequirement int64
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
