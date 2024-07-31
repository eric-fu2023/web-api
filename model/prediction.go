package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
) 

type Prediction struct {
	ploutos.Predictions
}

type ListPredictionCond struct {
	Page int
	Limit int
	AnalystId int64
}

func ListPredictions(cond ListPredictionCond) (preds []Prediction, err error) {

	// TODO : filter status 
	db := DB.
		Preload("Analyst").
		Scopes(Paginate(cond.Page, cond.Limit)).
		Where("deleted_at IS NULL")

	if (cond.AnalystId != 0) {
		db.Where("analyst_id", cond.AnalystId)
	}
	
	err = db.Find(&preds).Error
		
	return 
}