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