package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

func GetDollarJackpotByStreamerId(streamerId int64) (jackpot ploutos.DollarJackpot, err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Where("streamer_id = ?", streamerId).Where("status = ?", 1).First(&jackpot).Error
		return
	})

	return
}
