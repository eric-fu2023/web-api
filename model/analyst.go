package model

import (
	"context"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

type Analyst struct {
	ploutos.Analyst
}

func (Analyst) List(page, limit int) (list []Analyst, err error) {
	db := DB.Scopes(Paginate(page, limit))

	err = db.
		Preload("AnalystSource").
		Where("is_active", true).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Order("id DESC").
		Find(&list).Error
	return
}

func (Analyst) GetDetail(id int) (target Analyst, err error) {
	db := DB.Where("id", id)
	err = db.
		Preload("Prediction").
		Where("is_active", true).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Order("id DESC").
		First(&target).Error
	return
}

// func GetAnalyst(c context.Context, analystId int64) (analyst models.Analyst, err error) {
// 	err = DB.WithContext(c).Where("is_active").Where("id = ?", analystId).First(&analyst).Error
// 	return
// }

func GetFollowingAnalystList(c context.Context, userId int64, page, limit int) (followings []models.UserAnalystFollowing, err error) {
	err = DB.Preload("Analyst").WithContext(c).Where("user_id = ?", userId).Where("is_deleted = ?", false).Find(&followings).Scopes(Paginate(page, limit)).Error
	return
}

func GetFollowingAnalystStatus(c context.Context, userId, analystId int64) (following models.UserAnalystFollowing, err error) {
	err = DB.WithContext(c).Where("user_id = ?", userId).Where("analyst_id = ?", analystId).Limit(1).Find(&following).Error
	return
}

func UpdateUserFollowAnalystStatus(following models.UserAnalystFollowing) (err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Save(&following).Error
		return
	})

	return
}
