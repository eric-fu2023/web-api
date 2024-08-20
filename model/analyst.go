package model

import (
	"context"
	"errors"
	"log"
	"slices"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	fbService "blgit.rfdev.tech/taya/game-service/fb2/outcome_service"

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
		Preload("Predictions", AnalystPredictionFilter).
		Preload("Predictions.PredictionSelections").
		Preload("Predictions.PredictionSelections.FbOdds").
		Preload("Predictions.PredictionSelections.FbOdds.FbOddsOrderRequestList").
		Preload("Predictions.PredictionSelections.FbOdds.FbOddsOrderRequestList.TayaBetReport").
		Preload("Predictions.PredictionSelections.FbOdds.RelatedOdds", SortFbOddsByShortName).		
		Preload("Predictions.PredictionSelections.FbOdds.MarketGroupInfo").
		Where("is_active", true).
		Order("sort DESC")

	if fbSportId != 0 {
		db = db.
			Joins("join prediction_articles on prediction_articles.analyst_id = prediction_analysts.id").
			Group("prediction_analysts.id").
			Where("prediction_articles.fb_sport_id", fbSportId)
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
		Preload("Predictions", AnalystPredictionFilter).
		Preload("Predictions.PredictionSelections").
		Preload("Predictions.PredictionSelections.FbOdds").
		Preload("Predictions.PredictionSelections.FbOdds.FbOddsOrderRequestList").
		Preload("Predictions.PredictionSelections.FbOdds.FbOddsOrderRequestList.TayaBetReport").
		Preload("Predictions.PredictionSelections.FbOdds.RelatedOdds", SortFbOddsByShortName).		
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
		Preload("Analyst.Predictions", AnalystPredictionFilter).
		Preload("Analyst.Predictions.PredictionSelections").
		Preload("Analyst.Predictions.PredictionSelections.FbOdds").
		Preload("Analyst.Predictions.PredictionSelections.FbOdds.FbOddsOrderRequestList").
		Preload("Analyst.Predictions.PredictionSelections.FbOdds.FbOddsOrderRequestList.TayaBetReport").
		Preload("Analyst.Predictions.PredictionSelections.FbOdds.RelatedOdds", SortFbOddsByShortName).		
		Preload("Analyst.Predictions.PredictionSelections.FbOdds.MarketGroupInfo").
		Joins("JOIN prediction_analysts on prediction_analyst_followers.analyst_id = prediction_analysts.id").
		WithContext(c).
		Where("user_id = ?", userId).
		Where("prediction_analysts.is_active = ?", true).
		Order("updated_at desc").
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

func AnalystPredictionFilter (db *gorm.DB) *gorm.DB {
	db = db.Where("is_published", true).
			Order("published_at desc")
	return db 
}

func GetPredictionsFromAnalyst(analyst Analyst, sportId int) []Prediction {
	sortedPredictions := slices.Clone(analyst.Predictions)
	filteredSorted := []Prediction{}
	for _, p := range sortedPredictions {
		if (int64(sportId) == 0 || p.FbSportId == sportId) {
			filteredSorted = append(filteredSorted, p)
		}
	}

	slices.SortFunc(filteredSorted, func(a, b Prediction) int {
		return b.PublishedAt.Compare(a.PublishedAt) // newest to oldest 
	})
	return filteredSorted
}

func GetOutcomesFromPredictions(predictions []Prediction) []fbService.PredictionOutcome {
	outcomes := []fbService.PredictionOutcome{}
	for _, pred := range predictions {
		predDao := GetPredictionFromPrediction(pred)
		res, err := fbService.ComputePredictionOutcomesByOrderReport(predDao)

		if err != nil {
			log.Printf("Error computing prediction, %s\n", err.Error())
		}
		
		outcomes = append(outcomes, res)
	}
	return outcomes
}