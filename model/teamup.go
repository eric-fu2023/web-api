package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

type Teamup struct {
	ploutos.TipsAnalyst

	Predictions []ploutos.TipsAnalystPrediction `gorm:"foreignKey:AnalystId;references:ID"`
	Followers   []ploutos.UserAnalystFollowing  `gorm:"foreignKey:AnalystId;references:ID"`
}

func StartTeamup(following UserAnalystFollowing) (err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Save(&following).Error
		return
	})

	return
}
