package model

// import (
// 	"context"
// 	"web-api/model"
// )

// func AnalystList(c context.Context, page int, limit int, analystId int64) (list []models.Analyst, err error) {
// 	// Get Analyst List By Pagination
// 	err = DB.WithContext(c).Where("is_active").Scopes(model.Paginate(page, limit)).Order("created desc").Find(&list).Error
// 	return
// }

// func GetAnalyst(c context.Context, analystId int64) (analyst models.Analyst, err error) {
// 	// Get Analyst List By Pagination
// 	err = DB.WithContext(c).Where("is_active").Where("id = ?", analystId).First(&analyst).Error
// 	return
// }

// func GetFollowingAnalystList(c context.Context, analystId int64) (analyst models.Analyst, err error) {
// 	// Get Analyst List By Pagination
// 	err = DB.WithContext(c).Where("is_active").Where("id = ?", analystId).First(&analyst).Error
// 	return
// }
