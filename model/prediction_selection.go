package model

import (
	ploutosFB "blgit.rfdev.tech/taya/ploutos-object/fb/model"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)


type PredictionSelection struct {
	ploutos.PredictionsSelection

	FbOdds ploutosFB.FbOdds `gorm:"foreignKey:SelectionId;references:ID"`
	FbMatch ploutosFB.FbMatch `gorm:"foreignKey:MatchId;references:MatchID"`
}