package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"database/sql"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"
)

var (
	ErrInvalidReferralCode = errors.New("invalid referral code")
)

type UserReferral struct {
	ploutos.UserReferral

	Referral          *User              `gorm:"foreignKey:ReferralId;references:ID"`
	ReferralVipRecord *ploutos.VipRecord `gorm:"foreignKey:ReferralId;references:UserID"`
}

func LinkReferralWithDB(tx *gorm.DB, referralId int64, referralCode string) error {
	var referrer User
	err := tx.Table(referrer.TableName()).Where("referral_code = ?", referralCode).First(&referrer).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrInvalidReferralCode
	} else if err != nil {
		return fmt.Errorf("get referrer: %w", err)
	}

	userReferral := ploutos.UserReferral{
		ReferrerId: referrer.ID,
		ReferralId: referralId,
	}
	err = tx.Create(&userReferral).Error
	if err != nil {
		return fmt.Errorf("create user referral: %w", err)
	}

	return nil
}

type UserReferralCond struct {
	ReferrerIds []int64
	ReferralIds []int64
}

func GetUserReferrals(cond UserReferralCond) ([]ploutos.UserReferral, error) {
	var userReferrals []ploutos.UserReferral
	db := DB.Table(ploutos.UserReferral{}.TableName())
	if len(cond.ReferrerIds) > 0 {
		db = db.Where("referrer_id IN ?", cond.ReferrerIds)
	}
	if len(cond.ReferralIds) > 0 {
		db = db.Where("referral_id IN ?", cond.ReferralIds)
	}

	err := db.Find(&userReferrals).Error
	return userReferrals, err
}

func GetReferralCount(referrerId int64) (int64, error) {
	var count int64
	err := DB.Table(ploutos.UserReferral{}.TableName()).Where("referrer_id = ?", referrerId).Count(&count).Error
	return count, err
}

type GetReferralDetailsCond struct {
	ReferrerId    int64
	JoinTimeStart sql.NullTime
	JoinTimeEnd   sql.NullTime
}

func GetReferralDetails(cond GetReferralDetailsCond) ([]UserReferral, error) {
	var referrals []UserReferral
	db := DB.Table(ploutos.UserReferral{}.TableName())

	if cond.ReferrerId != 0 {
		db = db.Where("referrer_id = ?", cond.ReferrerId)
	}

	db = db.Preload("Referral").Joins("INNER JOIN users ON users.id = user_referrals.referral_id")
	if cond.JoinTimeStart.Valid {
		db = db.Where("users.created_at >= ?", cond.JoinTimeStart)
	}
	if cond.JoinTimeEnd.Valid {
		db = db.Where("users.created_at < ?", cond.JoinTimeEnd.Time.Add(time.Second)) // Postgres stores time with sub second precision
	}

	err := db.Table(ploutos.UserReferral{}.TableName()).
		Preload("ReferralVipRecord").
		Order("users.created_at desc").
		Find(&referrals).Error
	return referrals, err
}
