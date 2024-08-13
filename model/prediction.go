package model

import (
	"errors"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

type Prediction struct {
	ploutos.TipsAnalystPrediction

	PredictionSelections []PredictionSelection `gorm:"foreignKey:PredictionId;references:ID"`
	AnalystDetail        Analyst               `gorm:"foreignKey:AnalystId;references:ID"`
}

type ListPredictionCond struct {
	Page      int
	Limit     int
	AnalystId int64
	FbMatchId int64
	SportId   int64
}

func preloadPredictions() *gorm.DB {
	return DB.
		Preload("PredictionSelections").
		Preload("PredictionSelections.FbOdds").
		Preload("PredictionSelections.FbOdds.RelatedOdds").
		Preload("PredictionSelections.FbMatch").
		Preload("AnalystDetail").
		Preload("AnalystDetail.PredictionSource").
		Preload("PredictionSelections.FbMatch.HomeTeam").
		Preload("PredictionSelections.FbMatch.AwayTeam")
	// Preload("AnalystDetail.Followers").
	// Preload("AnalystDetail.Predictions").
}

func ListPredictions(cond ListPredictionCond) (preds []Prediction, err error) {

	db := preloadPredictions()
	db = db.
		Scopes(Paginate(cond.Page, cond.Limit)).
		Where("tips_analyst_predictions.deleted_at IS NULL").
		Where("tips_analyst_predictions.is_active", true).
		Joins("left join tips_analyst_prediction_selections on tips_analyst_prediction_selections.prediction_id = tips_analyst_predictions.id").
		Joins("left join fb_matches on tips_analyst_prediction_selections.match_id = fb_matches.match_id").
		Group("tips_analyst_predictions.id")

	if cond.AnalystId != 0 {
		db = db.Where("tips_analyst_predictions.analyst_id", cond.AnalystId)
	}

	if cond.FbMatchId != 0 {
		db = db.Where("tips_analyst_prediction_selections.match_id = ?", cond.FbMatchId)
	}

	if cond.SportId != 0 {
		db = db.Where("fb_matches.sports_id = ?", cond.SportId)
	}

	err = db.Find(&preds).Error

	return
}

func GetPrediction(predictionId int64) (pred Prediction, err error) {
	err = preloadPredictions().
		Where("deleted_at IS NULL").
		Where("is_active", true).
		Where("id = ?", predictionId).
		First(&pred).Error
	return
}

func PredictionExist(predictionId int64) (exist bool, err error) {
	err = DB.Where("id", predictionId).First(&Prediction{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func GetPredictionBetReports(predictionId int64) (reports []ploutos.FbBetReport, err error) {
	var pred Prediction

	err = DB.
		Preload("PredictionSelections").
		Preload("PredictionSelections.FbOdds").
		Preload("PredictionSelections.FbOdds.FbOddsOrderRequest").
		Preload("PredictionSelections.FbOdds.FbOddsOrderRequest.FbBetReport").
		Where("id", predictionId).
		First(&pred).
		Error

	if err != nil {
		return
	}

	for _, selection := range pred.PredictionSelections {
		if selection.FbOdds.FbOddsOrderRequest.FbBetReport.ID != nil {
			reports = append(reports, selection.FbOdds.FbOddsOrderRequest.FbBetReport)
		}
	}

	return
}
