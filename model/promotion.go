package model

import (
	"context"
	"time"

	models "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

type PromotionReward struct {
	Rewards []PromotionRewardDetail `json:"rewards"`
}

type PromotionRewardDetail struct {
	Rewards    []RewardDetail    `json:"rewards"`
	Conditions []RewardCondition `json:"conditions"`
}

type RewardDetail struct {
	Max      int64  `json:"max"`
	Type     string `json:"type"`
	Value    int64  `json:"value"`
	ValueMin string `json:"value_min"`
	ValueMax string `json:"value_max"`
}

type RewardCondition struct {
	Value          string `json:"value"`
	Operator       string `json:"operator"`
	ValueType      string `json:"value_type"`
	ReferenceValue int64  `json:"reference_value"`
}

type MissionTier struct {
	MissionAmount int64 `json:"mission_amount"`
	RewardAmount  int64 `json:"reward_amount"`
}

func OngoingPromotions(c context.Context, brandId int, now time.Time) (list []models.Promotion, err error) {
	err = DB.WithContext(c).Where("brand_id = ? or brand_id = 0", brandId).Where("is_active").Not("is_hide").Scopes(Ongoing(now, "start_at", "end_at")).Order("sort_factor desc").Find(&list).Error
	return
}

// OngoingPromotionById
// select scope should be identical to OngoingPromotions() ? idk.
func OngoingPromotionById(c context.Context, brandId int, promotionId int64, now time.Time) (p models.Promotion, err error) {
	err = DB.Debug().WithContext(c).Where("brand_id = ? or brand_id = 0", brandId).Where("is_active").Where("id", promotionId).Scopes(Ongoing(now, "start_at", "end_at")).First(&p).Error
	return
}

func PromotionGetActiveNoCheckStartEnd(c context.Context, brandID int, promotionID int64, now time.Time) (p models.Promotion, err error) {
	err = DB.Debug().WithContext(c).Where("brand_id = ? or brand_id = 0", brandID).Where("is_active").Where("id", promotionID).First(&p).Error
	return
}

func PromotionGetSubActive(c context.Context, brandID int, promotionID int64, now time.Time) (p []models.Promotion, err error) {
	err = DB.Debug().WithContext(c).Where("brand_id = ? or brand_id = 0", brandID).Where("is_active").Where("parent_id", promotionID).Find(&p).Error
	return
}

func PromotionGetActiveNoBrand(c context.Context, promotionID int64, now time.Time) (p models.Promotion, err error) {
	err = DB.Debug().WithContext(c).Where("is_active").Where("id", promotionID).Scopes(Ongoing(now, "start_at", "end_at")).First(&p).Error
	return
}

func GetPromotion(c context.Context, promotionID int64) (p models.Promotion, err error) {
	err = DB.Debug().WithContext(c).Where("id", promotionID).First(&p).Error
	return
}

func PromotionGetActivePassive(c context.Context, brandID int, now time.Time) (p []models.Promotion, err error) {
	err = DB.Debug().WithContext(c).Joins("JOIN promotion_sessions ON promotion_sessions.promotion_id = promotions.id AND promotion_sessions.start_at < ? AND promotion_sessions.end_at > ?", now, now).
		Where("brand_id = ? or brand_id = 0", brandID).Where("is_active").Where("type in ?", models.PassivePromotionType()).Scopes(Ongoing(now, "start_at", "end_at")).Find(&p).Error
	return
}

// Find Success ONLY entry (Return Id = 0 if Rejected / Pending)
// func FindJoinCustomPromotionEntry(c context.Context, brandID int, promotionID int64) (entry models.PromotionRequest, err error) {
// 	err = DB.WithContext(c).Where("brand_id = ? or brand_id = 0", brandID).Where("is_active").Not("is_hide").Where("status = 2").First(&entry).Error
// 	return
// }

func CreateJoinCustomPromotion(request models.PromotionRequest) (err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Create(&request).Error
		return
	})
	return
}

func CheckIfCustomPromotionEntryExceededLimit(c context.Context, entryLimitType, promotionId, userId int64, x int) (isExceeded bool, err error) {
	var list []models.PromotionRequest
	err = DB.Debug().WithContext(c).Where("status = 2 OR status = 1").Where("promotion_id", promotionId).Where("user_id", userId).Scopes(CustomPromotionEntryLimit(entryLimitType)).Find(&list).Error

	if err != nil || len(list) >= x {
		isExceeded = true
	}

	return
}

func CustomPromotionEntryLimit(entryLimitType int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		now := time.Now().UTC()

		switch entryLimitType {

		// case models.CustomPromotionClickLimit:
		// 	// Find Promotion Request to Check Total Entry Limit
		// 	return db.Where("promotion_id = ?", promotionId)
		case models.CustomPromotionClickDailyLimit:
			// Find Promotion Request to Check Current Day Entry Limit
			today := now.Format("2006-01-02")
			return db.Where("DATE(created_at) = ?", today)
		case models.CustomPromotionClickWeeklyLimit:
			// Find Promotion Request to Check Current Week Entry Limit
			_, week := now.ISOWeek()
			return db.Where("WEEK(created_at, 1) = ?", week)
		case models.CustomPromotionClickMonthlyLimit:
			// Find Promotion Request to Check Current Month Entry Limit
			month := now.Format("2006-01")
			return db.Where("DATE_FORMAT(created_at, '%Y-%m') = ?", month)
		default:
			// No Constraint
			return db
		}
	}
}
