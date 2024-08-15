package model

import (
	"context"
	"errors"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

type Analyst struct {
	ploutos.TipsAnalyst

	Predictions []Prediction                   `gorm:"foreignKey:AnalystId;references:ID"`
	Followers   []ploutos.UserAnalystFollowing `gorm:"foreignKey:AnalystId;references:ID"`
	AnalystSport 		AnalystSport 					`gorm:"foreignKey:ID;references:AnalystId"`
}

type AnalystSport struct {
	ploutos.AnalystSport 

	Sport []ploutos.SportType `gorm:"foreignKey:ID;references:SportId"`
}

func (Analyst) List(page, limit int, fbSportId int64) (list []Analyst, err error) {
	db := DB.Scopes(Paginate(page, limit))

	db = db.
		Preload("PredictionSource").
		Preload("Followers").
		Preload("Predictions").
		Preload("Predictions.PredictionSelections").
		Preload("Predictions.PredictionSelections.FbOdds").
		Preload("Predictions.PredictionSelections.FbOdds.RelatedOdds").
		Where("is_active", true).
		Order("created_at DESC").
		Order("id DESC")

	if fbSportId != 0 {
		db = db.
		Joins("JOIN analyst_sport ON analyst_sport.analyst_id = tips_analysts.id").
		Joins("JOIN sport_type ON analyst_sport.sport_id = sport_type.id").
		Where("fb_sport_id = ?", fbSportId)
	}

	err = db.
		Find(&list).
		Error

	return
}

func (Analyst) GetDetail(id int) (target Analyst, err error) {
	db := DB.Where("id", id)
	err = db.
		Preload("PredictionSource").
		Preload("Predictions").
		Preload("Predictions.PredictionSelections").
		Preload("Predictions.PredictionSelections.FbOdds").
		Preload("Predictions.PredictionSelections.FbOdds.RelatedOdds").
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

func AnalystExist(analystId int64) (exist bool, err error) {
	err = DB.Where("id", analystId).First(&Analyst{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
