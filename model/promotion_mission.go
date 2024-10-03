package model

import (
	"context"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

func GetMissionByPromotionId(c context.Context, brandId int, promotionId int64) (list []models.PromotionMission, err error) {
	err = DB.WithContext(c).Where("promotion_id = ?", promotionId).Where("is_active").Order("weightage desc").Find(&list).Error
	return
}

func GetMissionById(c context.Context, brandId int, promotionMissionId int64) (mission models.PromotionMission, err error) {
	err = DB.WithContext(c).Where("id = ?", promotionMissionId).Where("is_active").First(&mission).Error
	return
}
