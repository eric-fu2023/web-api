package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

const ContributionLimitPercent = 0.5

type DollarJackpotDraw struct {
	ploutos.DollarJackpotDraw
	Total  *int64 `gorm:"-"`
	Winner User   `gorm:"references:WinnerId;foreignKey:ID"`
}

type ContributionSum struct {
	Sum int64
}

func GetContribution(userId int64, drawId int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`user_id`, userId).Where(`game_id`, drawId).Select(`SUM(bet) as sum`)
	}
}
func GetTotalContribution(drawId int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`game_id`, drawId).Select(`SUM(bet) as sum`)
	}
}