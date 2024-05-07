package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"database/sql"
	"gorm.io/gorm"
	"time"
)

type ReferralAllianceSummary struct {
	ReferrerId      int64
	ReferralId      int64
	RecordCount     int64
	TotalReward     int64
	ClaimableReward int64
}

type GetReferralAllianceSummaryCond struct {
	ReferrerIds    []int64
	ReferralIds    []int64
	HasBeenClaimed []bool
	RewardMonthEnd string
}

func GetReferralAllianceSummaries(cond GetReferralAllianceSummaryCond) ([]ReferralAllianceSummary, error) {
	db := DB.Table(ploutos.ReferralAllianceReward{}.TableName())
	selectFields := []string{"COUNT(*) AS record_count", "SUM(amount) AS total_reward", "SUM(claimable_amount) AS claimable_reward"}

	if len(cond.ReferrerIds) > 0 {
		db = db.Where("referrer_id IN ?", cond.ReferrerIds)
		db = db.Group("referrer_id")
		selectFields = append(selectFields, "referrer_id")
	}
	if len(cond.ReferralIds) > 0 {
		db = db.Where("referral_id IN ?", cond.ReferralIds)
		db = db.Group("referral_id")
		selectFields = append(selectFields, "referral_id")
	}
	if len(cond.HasBeenClaimed) > 0 {
		db = db.Where("has_been_claimed IN ?", cond.HasBeenClaimed)
	}
	if cond.RewardMonthEnd != "" {
		db = db.Where("reward_month <= ?", cond.RewardMonthEnd)
	}

	var res []ReferralAllianceSummary
	err := db.Table(ploutos.ReferralAllianceReward{}.TableName()).
		Select(selectFields).
		Where("reward_month != ''"). // filter out old data TODO remove this after a while
		Find(&res).Error
	return res, err
}

type GetReferralAllianceRewardsCond struct {
	ReferrerIds      []int64
	ReferralIds      []int64
	HasBeenClaimed   []bool
	RewardMonthStart string
	RewardMonthEnd   string
}

func GetReferralAllianceRewards(cond GetReferralAllianceRewardsCond) ([]ploutos.ReferralAllianceReward, error) {
	db := DB.Table(ploutos.ReferralAllianceReward{}.TableName())

	if len(cond.ReferrerIds) > 0 {
		db = db.Where("referrer_id IN ?", cond.ReferrerIds)
	}
	if len(cond.ReferralIds) > 0 {
		db = db.Where("referral_id IN ?", cond.ReferralIds)
	}
	if len(cond.HasBeenClaimed) > 0 {
		db = db.Where("has_been_claimed IN ?", cond.HasBeenClaimed)
	}
	if cond.RewardMonthStart != "" {
		db = db.Where("reward_month >= ?", cond.RewardMonthStart)
	}
	if cond.RewardMonthEnd != "" {
		db = db.Where("reward_month <= ?", cond.RewardMonthEnd)
	}

	var rewards []ploutos.ReferralAllianceReward
	err := db.
		Where("reward_month != ''"). // filter out old data TODO remove this after a while
		Order("reward_month DESC").
		Find(&rewards).Error
	return rewards, err
}

func ClaimReferralAllianceRewards(tx *gorm.DB, ids []int64, now time.Time) error {
	return tx.Table(ploutos.ReferralAllianceReward{}.TableName()).
		Where("id IN ?", ids).
		Updates(ploutos.ReferralAllianceReward{
			HasBeenClaimed: true,
			ClaimTime:      sql.NullTime{Time: now, Valid: true}},
		).Error
}
