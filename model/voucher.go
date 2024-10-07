package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func AmountReplace(original string, amount float64) string {
	return util.TextReplace(original, map[string]string{
		"amount": fmt.Sprintf("%f", amount),
	})
}

func VoucherGetByUniqueID(c context.Context, uniqueID string) (v models.Voucher, err error) {
	err = DB.WithContext(c).Where("unique_id", uniqueID).First(&v).Error
	return
}

func GetVoucherByUserAndPromotionSession(c context.Context, userID int64, promotionSessionID int64) (v models.Voucher, err error) {
	err = DB.Debug().WithContext(c).Where("user_id", userID).Where("promotion_session_id", promotionSessionID).Order("created_at desc").First(&v).Error
	return
}

func VoucherListUsableByUserFilter(c context.Context, userID int64, filter string, now time.Time) (v []models.Voucher, err error) {
	db := DB.Debug().WithContext(c).Where("user_id", userID).Where("is_usable")
	switch filter {
	default:
		err = InvalidFilter()
		return
	case "all":
	case "used":
		db = db.Where("status IN ?", []int{models.VoucherStatusRedeemed, models.VoucherStatusPending})
	case "expired":
		db = db.Where("status not IN ?", []int{models.VoucherStatusRedeemed, models.VoucherStatusPending}).Where("end_at < ?", time.Now())
	case "valid":
		db = db.Where("status", models.VoucherStatusReady).Scopes(Ongoing(time.Now(), "start_at", "end_at"))
	}
	err = db.Order("id desc").Find(&v).Error
	return
}

func VoucherActiveGetByIDUserWithDB(c context.Context, userID int64, ID int64, now time.Time, tx *gorm.DB) (v models.Voucher, err error) {
	err = tx.Debug().WithContext(c).Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id", userID).Where("id", ID).Scopes(Ongoing(now, "start_at", "end_at")).First(&v).Error
	return
}

func VoucherPendingGetByIDUserWithDB(c context.Context, userID int64, ID int64, now time.Time, tx *gorm.DB) (v models.Voucher, err error) {
	err = tx.Debug().WithContext(c).Clauses(clause.Locking{Strength: "UPDATE"}).Where("status", models.VoucherStatusPending).Where("user_id", userID).Where("id", ID).Scopes(Ongoing(now, "start_at", "end_at")).First(&v).Error
	return
}

func VoucherPendingGetByIDUserWithDBNoTime(c context.Context, userID int64, ID int64, tx *gorm.DB) (v models.Voucher, err error) {
	err = tx.Debug().WithContext(c).Clauses(clause.Locking{Strength: "UPDATE"}).Where("status", models.VoucherStatusPending).Where("user_id", userID).Where("id", ID).First(&v).Error
	return
}

func InvalidFilter() error {
	return errors.New("invalid_filter")
}

func GetVoucherByUserAndPromotionAndReference(c context.Context, userId, promotionId int64, referenceId string) (v models.Voucher, err error) {
	err = DB.Debug().WithContext(c).Where("user_id", userId).Where("promotion_id", promotionId).Where("reference_id", referenceId).First(&v).Error
	return
}

func GetVouchersByUserAndPromotion(c context.Context, userId, promotionId int64) (v []models.Voucher, err error) {
	err = DB.Debug().WithContext(c).Where("user_id", userId).Where("promotion_id", promotionId).Find(&v).Error
	return
}
