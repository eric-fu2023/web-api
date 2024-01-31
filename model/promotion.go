package model

import (
	"context"
	"time"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

func PromotionList(c context.Context, brandID int, now time.Time) (list []models.Promotion, err error) {
	err = DB.WithContext(c).Where("brand_id", brandID).Scopes(Ongoing(now, "start_at", "end_at")).Find(&list).Error
	return
}

func PromotionGetActive(c context.Context, brandID int, promotionID int64, now time.Time) (p models.Promotion, err error) {
	err = DB.WithContext(c).Where("id", promotionID).Scopes(Ongoing(now, "start_at", "end_at")).First(&p).Error
	return
}
