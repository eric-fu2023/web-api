package model

import (
	"context"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

type Analyst struct {
	ploutos.TipsAnalyst

	Predictions []Prediction `gorm:"foreignKey:AnalystId;references:ID"`
	Followers []ploutos.UserAnalystFollowing `gorm:"foreignKey:AnalystId;references:ID"`
}

func (Analyst) List(page, limit int, sportId int64) (list []Analyst, err error) {
	db := DB.Scopes(Paginate(page, limit))

	db = db.
		Preload("TipsAnalystSource").
		Preload("Followers").
		Preload("Predictions").
		Where("is_active", true).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Order("id DESC").

	if sportId != 0 {
		// TODO : Add filter for sport id when analyst struct finalised in backend API
	}


	err = db.
		Find(&list).
		Error
	
	return
}

func (Analyst) GetDetail(id int) (target Analyst, err error) {
	db := DB.Where("id", id)
	err = db.
		Preload("TipsAnalystSource").
		Preload("Predictions").
		Preload("Followers").
		Where("is_active", true).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Order("id DESC").
		First(&target).Error
	return
}

func GetFollowingAnalystList(c context.Context, userId int64, page, limit int) (followings []UserAnalystFollowing, err error) {
	err = DB.Preload("Analyst").
	Preload("Analyst.Followers").
	Preload("Analyst.Predictions").
	WithContext(c).
	Where("user_id = ?", userId).Where("is_deleted = ?", false).Find(&followings).Scopes(Paginate(page, limit)).Error
	return
}

func GetFollowingAnalystStatus(c context.Context, userId, analystId int64) (following UserAnalystFollowing, err error) {
	err = DB.WithContext(c).Where("user_id = ?", userId).Where("analyst_id = ?", analystId).Limit(1).Find(&following).Error
	return
}

func UpdateUserFollowAnalystStatus(following UserAnalystFollowing) (err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Save(&following).Error
		return
	})

	return
}
