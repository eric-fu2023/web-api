package model

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

type Match models.MatchC

func MatchWithDetails(db *gorm.DB) *gorm.DB {
	return db.Preload(`Competition`).
		Preload(`Home`).
		Preload(`Away`)
}
