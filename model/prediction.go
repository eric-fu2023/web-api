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

func ListPredictions(cond ListPredictionCond) (preds []Prediction, err error) {

	// TODO : filter status
	db := DB.
		Preload("PredictionSelections").
		Preload("PredictionSelections.FbOdds").
		Preload("PredictionSelections.FbOdds.RelatedOdds").
		Preload("PredictionSelections.FbMatch").
		Preload("AnalystDetail").
		Preload("AnalystDetail.PredictionSource").
		Preload("PredictionSelections.FbMatch.HomeTeam").
		Preload("PredictionSelections.FbMatch.AwayTeam").
		// Preload("AnalystDetail.Followers").
		// Preload("AnalystDetail.Predictions").
		Scopes(Paginate(cond.Page, cond.Limit)).
		Where("tips_analyst_predictions.deleted_at IS NULL").
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
	err = DB.
		Debug().
		Preload("PredictionSelections").
		Preload("PredictionSelections.FbOdds").
		Preload("PredictionSelections.FbOdds.RelatedOdds").
		Preload("PredictionSelections.FbMatch").
		Preload("AnalystDetail").
		Preload("AnalystDetail.PredictionSource").
		Preload("PredictionSelections.FbMatch.HomeTeam").
		Preload("PredictionSelections.FbMatch.AwayTeam").
		// Preload("AnalystDetail.Followers").
		// Preload("AnalystDetail.Predictions").
		Where("deleted_at IS NULL").
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
