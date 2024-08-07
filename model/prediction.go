package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
) 

type Prediction struct {
	ploutos.TipsAnalystPrediction

	PredictionSelections []PredictionSelection `gorm:"foreignKey:PredictionId;references:ID"`
	AnalystDetail Analyst `gorm:"foreignKey:AnalystId;references:ID"`
}

type ListPredictionCond struct {
	Page int
	Limit int
	AnalystId int64
	FbMatchId int64
	SportId int64
}

func ListPredictions(cond ListPredictionCond) (preds []Prediction, err error) {

	// TODO : filter status 
	db := DB.
		Preload("PredictionSelections").
		Preload("PredictionSelections.FbOdds").
		Preload("PredictionSelections.FbMatch").
		Preload("AnalystDetail").
		Preload("AnalystDetail.TipsAnalystSource").
		// Preload("AnalystDetail.Followers").
		// Preload("AnalystDetail.Predictions").
		Scopes(Paginate(cond.Page, cond.Limit)).
		Where("tips_analyst_predictions.deleted_at IS NULL")

	if (cond.AnalystId != 0) {
		db = db.Where("analyst_id", cond.AnalystId)
	}

	if (cond.FbMatchId != 0) {
		db = db.Joins("join tips_analyst_prediction_selections on tips_analyst_prediction_selections.prediction_id = tips_analyst_predictions.id").
		Where("tips_analyst_prediction_selections.match_id = ?", cond.FbMatchId).
		Group("tips_analyst_predictions.id")
	}

	if (cond.SportId != 0) {
		db = db.Joins("join tips_analyst_prediction_selections on tips_analyst_prediction_selections.prediction_id = tips_analyst_predictions.id").
		Joins("join fb_matches on tips_analyst_prediction_selections.match_id = fb_matches.match_id").
		Where("fb_matches.sports_id = ?", cond.SportId).
		Group("tips_analyst_predictions.id")
	}
	
	err = db.Find(&preds).Error
		
	return 
}

func GetPrediction(predictionId int64) (pred Prediction, err error) {
	err = DB.
		Preload("PredictionSelections").
		Preload("PredictionSelections.FbOdds").
		Preload("PredictionSelections.FbMatch").
		Preload("AnalystDetail").
		Preload("AnalystDetail.TipsAnalystSource").
		// Preload("AnalystDetail.Followers").
		// Preload("AnalystDetail.Predictions").
		Where("deleted_at IS NULL").
		Where("id = ?", predictionId).
		First(&pred).Error
	return
}