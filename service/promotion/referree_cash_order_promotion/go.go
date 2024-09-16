package referree_cash_order_promotion

import (
	"context"
	"errors"
	"gorm.io/gorm"

	po "blgit.rfdev.tech/taya/ploutos-object"
)

// Service
// To be called after relevant records (cash order, transaction, cash method etc) are updated in the database
type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) (*Service, error) {
	if db == nil {
		return nil, errors.New("no gorm")
	}
	return &Service{}, nil
}

// AddRewardForClosedDeposit
// adds redeemable reward for the referrer
//
// Example
// 1. A refers B
// 2. B deposits (1st, 2nd....)
// 3. Server add reward for A
func (s *Service) AddRewardForClosedDeposit(ctx context.Context, referree UserForm, referreeCashOrder int64) error {
	var referreeDbInfo po.User
	db := s.db.Where(`username`, referree.Id)
	if err := db.Scopes(po.ByActiveNonStreamerUser).First(&referreeDbInfo).Error; err != nil {
		return err
	}

	var referrer UserReferrer

	rErr := db.Debug().Table("user_referrals").
		Joins("LEFT JOIN users ON users.id = user_referrals.referral_id").
		Where("user_referrals.referral_id = ?", referree.Id).
		Select("users.id as user_id, user_referrals.referral_id as referral_id").Find(&referrer).Error

	if rErr != nil {
		return rErr
	}

	return nil
}

// AddRewardForClosedDeposit
// adds redeemable reward for the referrer
//
// Example
// 1. A refers B
// 2. B deposits (1st, 2nd....)
// 3. Server add reward for A
func (s *Service) GetRewardsFor(ctx context.Context, referree UserForm, referreeCashOrder int64) error {
	var referreeDbInfo po.User
	db := s.db.Where(`username`, referree.Id)
	if err := db.Scopes(po.ByActiveNonStreamerUser).First(&referreeDbInfo).Error; err != nil {
		return err
	}

	var referrer UserReferrer

	rErr := db.Debug().Table("user_referrals").
		Joins("LEFT JOIN users ON users.id = user_referrals.referral_id").
		Where("user_referrals.referral_id = ?", referree.Id).
		Select("users.id as user_id, user_referrals.referral_id as referral_id").Find(&referrer).Error

	if rErr != nil {
		return rErr
	}

	return nil
}

type _ = po.UserReferral
type _ = po.User

type UserReferrer struct {
	Id         int64 `gorm:"primarykey" gorm:"id"` // 主键ID
	UserId     int64 `gorm:"column:user_id;"`
	ReferrerId int64 `gorm:"column:referral_id;"`
}
