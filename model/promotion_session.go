package model

import (
	"context"
	"time"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

type PromotionSession struct {
	models.PromotionSession
}

func PromotionSessionGetActive(c context.Context, promotionID int64, now time.Time) (p PromotionSession, err error) {
	err = DB.WithContext(c).Where("promotion_id", promotionID).Scopes(Ongoing(now, "start_at", "end_at")).First(&p).Error
	return
}
