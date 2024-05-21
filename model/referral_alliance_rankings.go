package model

import ploutos "blgit.rfdev.tech/taya/ploutos-object"

type ReferralAllianceRankingInfo struct {
	ReferrerId     int64
	TotalClaimable int64
	ReferralCount  int64

	Referrer *User `gorm:"foreignKey:ReferrerId;references:ID"`
}

func GetTopReferralAllianceRewardRankings(limit int) ([]ReferralAllianceRankingInfo, error) {
	var ret []ReferralAllianceRankingInfo

	rankings := DB.Table(ploutos.ReferralAllianceReward{}.TableName()).
		Select("referrer_id, SUM(claimable_amount) AS total_claimable").
		Group("referrer_id")

	topRankings := DB.Table("(?) rankings", rankings).
		Where("total_claimable > 0").
		Order("total_claimable DESC").
		Limit(limit)

	err := DB.Table("(?) top_rankings", topRankings).
		Preload("Referrer").
		Find(&ret).Error

	return ret, err
}

func GetTopReferralAllianceReferralRankings(limit int) ([]ReferralAllianceRankingInfo, error) {
	var ret []ReferralAllianceRankingInfo

	err := DB.Table(ploutos.UserReferral{}.TableName()).
		Preload("Referrer").
		Select("referrer_id, COUNT(*) AS referral_count").
		Group("referrer_id").
		Order("referral_count DESC").
		Limit(limit).
		Find(&ret).Error
	return ret, err
}

func GetUserReferralAllianceRankingInfo(userId int64) (ReferralAllianceRankingInfo, error) {
	var ret ReferralAllianceRankingInfo

	referralCount := DB.Table(ploutos.UserReferral{}.TableName()).
		Select("referrer_id, COUNT(*) AS referral_count").
		Where("referrer_id = ?", userId).
		Group("referrer_id")

	rewardAmount := DB.Table(ploutos.ReferralAllianceReward{}.TableName()).
		Select("referrer_id, SUM(claimable_amount) AS total_claimable").
		Where("referrer_id = ?", userId).
		Group("referrer_id")

	err := DB.Table("(?) rc", referralCount).
		Preload("Referrer").
		Joins("LEFT JOIN (?) ra ON rc.referrer_id = ra.referrer_id", rewardAmount).
		Select("rc.referrer_id, rc.referral_count, ra.total_claimable").
		Find(&ret).Error
	return ret, err
}
