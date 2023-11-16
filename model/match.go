package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

type Match ploutos.Match

func MatchWithDetails(db *gorm.DB) *gorm.DB {
	return db.Preload(`Competition`).
		Preload(`Home`).
		Preload(`Away`)
}
