package model

import (
	"context"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

type Analyst struct {
	ploutos.Analyst

	Followers []ploutos.UserAnalystFollowing `gorm:"foreignKey:AnalystId;references:ID"`
}

func (Analyst) List(page, limit int) (list []Analyst, err error) {
	db := DB.Scopes(Paginate(page, limit))

	err = db.
		Preload("AnalystSource").
		Preload("Followers").
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
		Preload("Predictions").
		Preload("Followers").
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

func GetFollowingAnalystList(c context.Context, userId int64, page, limit int) (followings []ploutos.UserAnalystFollowing, err error) {
	err = DB.Preload("Analyst").WithContext(c).Where("user_id = ?", userId).Where("is_deleted = ?", false).Find(&followings).Scopes(Paginate(page, limit)).Error
	return
}

func GetFollowingAnalystStatus(c context.Context, userId, analystId int64) (following ploutos.UserAnalystFollowing, err error) {
	err = DB.WithContext(c).Where("user_id = ?", userId).Where("analyst_id = ?", analystId).Limit(1).Find(&following).Error
	return
}

func UpdateUserFollowAnalystStatus(following ploutos.UserAnalystFollowing) (err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Save(&following).Error
		return
	})

	return
}
