package model

import (
	"context"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

// func AnalystList(c context.Context, page int, limit int, analystId int64) (list []models.Analyst, err error) {
// 	// Get Analyst List By Pagination
// 	err = DB.WithContext(c).Where("is_active").Scopes(model.Paginate(page, limit)).Order("created desc").Find(&list).Error
// 	return
// }

// func GetAnalyst(c context.Context, analystId int64) (analyst models.Analyst, err error) {
// 	err = DB.WithContext(c).Where("is_active").Where("id = ?", analystId).First(&analyst).Error
// 	return
// }

func GetFollowingAnalystList(c context.Context, userId int64, page, limit int) (followings []models.UserAnalystFollowing, err error) {
	err = DB.WithContext(c).Where("user_id = ?", userId).Scopes(Paginate(page, limit)).Find(&followings).Error
	return
}

func GetFollowingAnalystStatus(c context.Context, userId, analystId int64) (following models.UserAnalystFollowing, err error) {
	err = DB.WithContext(c).Where("user_id = ?", userId).Where("analyst_id = ?", analystId).First(&following).Error
	return
}

func UpdateUserFollowAnalystStatus(following models.UserAnalystFollowing) (err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Save(&following).Error
		return
	})

	return
}
