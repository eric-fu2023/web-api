package model

import (
	"context"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

func GetMissionByPromotionId(c context.Context, brandId int, promotionId int64) (list []models.PromotionMission, err error) {
	err = DB.WithContext(c).Where("promotion_id = ?", promotionId).Where("is_active").Order("weightage desc").Find(&list).Error
	return
}
