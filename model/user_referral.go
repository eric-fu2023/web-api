package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

var (
	ErrInvalidReferralCode = errors.New("invalid referral code")
)

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
