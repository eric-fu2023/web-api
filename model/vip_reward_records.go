package model

import (
	"errors"
	"time"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetVipRewardRecord(tx *gorm.DB, userID int64, promoID int64, after time.Time) (models.VipRewardRecords, error) {
	ret := models.VipRewardRecords{}
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id", userID).Where("promotion_id", promoID).Order("date").Where("date >= ?", after).First(&ret).Error
	return ret, err
}

func MarkVipRewardRecordsClaimed(tx *gorm.DB, id int64, cashOrderID string) error {
	if db := tx.Where("id", id).Where("status", 1).Updates(map[string]any{
		"status":        2,
		"cash_order_id": cashOrderID,
	}); db.RowsAffected != 1 {
		return errors.New("reward cannot be claimed")
	} else {
		return db.Error
	}
}
