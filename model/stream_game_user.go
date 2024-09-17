package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

const ()

func GetStreamGameStreamer(userId, gameId int64, gameType int64) (hasGame bool, err error) {
	var record ploutos.StreamGameUser
	err = DB.Transaction(func(tx *gorm.DB) error {
		tx = tx.Table("stream_game_users").
			Where("user_id = ?", userId).
			Where("game_type",gameType)
		if gameId!=0{
			tx = tx.Where("stream_game_id = ?", gameId)
		}
		if err := tx.First(&record).Error; err != nil {
			return err
		}
		return nil
	})

	if record.ID == 0 || !record.IsActive {
		return
	}

	hasGame = true
	return
}

func ToggleStreamerStreamGame(streamerId, gameId int64, status bool) (err error) {

	var game ploutos.StreamGameUser

	_ = DB.Table("stream_game_users").
		Where("user_id = ?", streamerId).
		Where("stream_game_id = ?", gameId).
		First(&game).Error

	if game.ID == 0 {
		game.UserId = streamerId
		game.StreamGameId = gameId
	}

	game.IsActive = status
	err = DB.Save(&game).Error

	return
}
