package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
) 

type Prediction struct {
	ploutos.Predictions

	PredictionSelections []ploutos.PredictionsSelection `gorm:"foreignKey:PredictionId;references:ID"`
}

type ListPredictionCond struct {
	Page int
	Limit int
	AnalystId int64
}

func ListPredictions(cond ListPredictionCond) (preds []Prediction, err error) {

	// TODO : filter status 
	db := DB.
		Scopes(Paginate(cond.Page, cond.Limit)).
		Where("deleted_at IS NULL")

	if (cond.AnalystId != 0) {
		db.Where("analyst_id", cond.AnalystId)
	}
	
	err = db.Find(&preds).Error
		
	return 
}

func GetPrediction(predictionId int64) (pred Prediction, err error) {
	err = DB.Preload("PredictionSelections").Where("deleted_at IS NULL").Where("id = ?", predictionId).First(&pred).Error
	return
}