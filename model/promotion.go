package model

import (
	"context"
	"time"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

func PromotionList(c context.Context, brandID int, now time.Time) (list []models.Promotion, err error) {
	err = DB.WithContext(c).Where("brand_id = ? or brand_id = 0", brandID).Where("is_active").Not("is_hide").Scopes(Ongoing(now, "start_at", "end_at")).Order("sort_factor desc").Find(&list).Error
	return
}

func PromotionGetActive(c context.Context, brandID int, promotionID int64, now time.Time) (p models.Promotion, err error) {
	err = DB.Debug().WithContext(c).Where("brand_id = ? or brand_id = 0", brandID).Where("is_active").Where("id", promotionID).Scopes(Ongoing(now, "start_at", "end_at")).First(&p).Error
	return
}

func PromotionGetActiveNoBrand(c context.Context, promotionID int64, now time.Time) (p models.Promotion, err error) {
	err = DB.Debug().WithContext(c).Where("is_active").Where("id", promotionID).Scopes(Ongoing(now, "start_at", "end_at")).First(&p).Error
	return
}

func PromotionGetActivePassive(c context.Context, brandID int, now time.Time) (p []models.Promotion, err error) {
	err = DB.Debug().WithContext(c).Joins("JOIN promotion_sessions ON promotion_sessions.promotion_id = promotions.id AND promotion_sessions.start_at < ? AND promotion_sessions.end_at > ?", now, now).
		Where("brand_id = ? or brand_id = 0", brandID).Where("is_active").Where("type in ?", models.PassivePromotionType()).Scopes(Ongoing(now, "start_at", "end_at")).Find(&p).Error
	return
}
