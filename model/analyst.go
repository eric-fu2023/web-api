package model

import (
	"context"
	"errors"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

type Analyst struct {
	ploutos.PredictionAnalyst

	Predictions  []Prediction `gorm:"foreignKey:AnalystId;references:ID"`
	AnalystSport AnalystSport `gorm:"foreignKey:ID;references:AnalystId"`
}

type AnalystSport struct {
	ploutos.AnalystSport

	Sport []ploutos.SportType `gorm:"foreignKey:ID;references:SportId"`
}

func (Analyst) List(page, limit int, fbSportId int64) (list []Analyst, err error) {
	db := DB.Scopes(Paginate(page, limit))

	db = db.
		Preload("PredictionAnalystSource").
		Preload("PredictionAnalystFollowers").
		Preload("Predictions", "is_published = ?", true).
		Preload("Predictions.PredictionSelections").
		Preload("Predictions.PredictionSelections.FbOdds").
		Preload("Predictions.PredictionSelections.FbOdds.FbOddsOrderRequestList").
		Preload("Predictions.PredictionSelections.FbOdds.FbOddsOrderRequestList.TayaBetReport").
		Preload("Predictions.PredictionSelections.FbOdds.RelatedOdds", func (db *gorm.DB) *gorm.DB  {
			/*
				order by
				CASE short_name_cn
					WHEN '和' THEN 2
					WHEN '主' THEN 1
					WHEN '客' THEN 3
					ELSE 0
				END,
				selection_type	
			*/
			db  = db.Order("CASE short_name_cn WHEN '和' THEN 2 WHEN '主' THEN 1 WHEN '客' THEN 3 ELSE 0 END, selection_type") 
			return db
		}).		
		Preload("Predictions.PredictionSelections.FbOdds.MarketGroupInfo").
		Where("is_active", true).
		Order("sort DESC")

	if fbSportId != 0 {
		db = db.
			Where("? = ANY(prediction_analysts.fb_sport_ids)", fbSportId)
	}

	err = db.
		Find(&list).
		Error

	return
}

func (Analyst) GetDetail(id int) (target Analyst, err error) {
	db := DB.Where("id", id)
	err = db.
		Preload("PredictionAnalystSource").
		Preload("Predictions", "is_published = ?", true).
		Preload("Predictions.PredictionSelections").
		Preload("Predictions.PredictionSelections.FbOdds").
		Preload("Predictions.PredictionSelections.FbOdds.FbOddsOrderRequestList").
		Preload("Predictions.PredictionSelections.FbOdds.FbOddsOrderRequestList.TayaBetReport").
		Preload("Predictions.PredictionSelections.FbOdds.RelatedOdds").
		Preload("Predictions.PredictionSelections.FbOdds.MarketGroupInfo").
		Preload("PredictionAnalystFollowers").
		Where("is_active", true).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Order("id DESC").
		First(&target).Error
	return
}

func GetFollowingAnalystList(c context.Context, userId int64, page, limit int) (followings []UserAnalystFollowing, err error) {
	err = DB.
		Scopes(Paginate(page, limit)).
		Preload("Analyst").
		Preload("Analyst.PredictionAnalystSource").
		Preload("Analyst.PredictionAnalystFollowers").
		Preload("Analyst.Predictions", "is_published = ?", true).
		Joins("JOIN prediction_analysts on prediction_analyst_followers.analyst_id = prediction_analysts.id").
		WithContext(c).
		Where("user_id = ?", userId).
		Where("prediction_analysts.is_active = ?", true).
		Find(&followings).Error
	return
}

func GetFollowingAnalystStatus(c context.Context, userId, analystId int64) (following UserAnalystFollowing, err error) {
	err = DB.Unscoped().WithContext(c).Where("user_id = ?", userId).Where("analyst_id = ?", analystId).Limit(1).Find(&following).Error
	return
}

func UpdateUserFollowAnalystStatus(following UserAnalystFollowing) (err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Save(&following).Error
		return
	})

	return
}

func SoftDeleteUserFollowAnalyst(following UserAnalystFollowing) (err error) {
	err = DB.Transaction(func(tx *gorm.DB) error {
		return tx.Delete(&following).Error
	})
	return
}

func RestoreUserFollowAnalyst(following UserAnalystFollowing) (err error) {
	err = DB.Transaction(func(tx *gorm.DB) error {
		return tx.Unscoped().Model(&following).Update("DeletedAt", nil).Error
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
