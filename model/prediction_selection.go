package model

import (
	// ploutosFB "blgit.rfdev.tech/taya/ploutos-object/fb/model"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)


type PredictionSelection struct {
	ploutos.PredictionArticleBet

	FbOdds FbOdds `gorm:"foreignKey:FbOddsId;references:ID"`
	FbMatch FbMatch `gorm:"foreignKey:FbMatchId;references:MatchID"`
}
