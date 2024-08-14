package model

import (
	// ploutosFB "blgit.rfdev.tech/taya/ploutos-object/fb/model"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)


type PredictionSelection struct {
	ploutos.TipsAnalystPredictionSelection

	FbOdds FbOdds `gorm:"foreignKey:SelectionId;references:ID"`
	FbMatch FbMatch `gorm:"foreignKey:MatchId;references:MatchID"`
}

// func GetSelectionBetReport(selectionId int64) (reports []ploutos.FbBetReport, err error) { //TODO remove when unused
// 	var selection PredictionSelection

// 	err = DB.
// 		Preload("FbOdds").
// 		Preload("FbOdds.FbOddsOrderRequest").
// 		Preload("FbOdds.FbOddsOrderRequest.FbBetReport").
// 		Where("id", selectionId).
// 		First(&selection).
// 		Error

// 	if err != nil {
// 		return 
// 	}
	
// 	for _, report := range selection.FbOdds.FbOddsOrderRequestList {
// 		reports = append(reports, report.FbBetReport)
// 	}

// 	return 
// }