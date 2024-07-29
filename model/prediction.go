package model

// import (
// 	"context"
// 	"web-api/model"
// )

// func PredictionListByAnalystId(c context.Context, page int, limit int, analystId int64) (list []models.Prediction, err error) {
// 	// Get Analyst Prediction List By Pagination
// 	err = DB.WithContext(c).Where("is_active").Where("analyst_id = ?", analystId).Scopes(model.Paginate(page, limit)).Order("created desc").Find(&list).Error
// 	return
// }
