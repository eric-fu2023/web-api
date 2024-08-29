package model

import (
	"web-api/conf/consts"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

type Streamer struct {
	ploutos.User
	IsLive   bool
	UserTags []ploutos.UserTag `gorm:"many2many:user_tag_conns;foreignKey:ID;joinForeignKey:UserId;References:ID;joinReferences:UserTagId"`
}

func StreamerWithLiveStream(includeUpcoming bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`users.role`, consts.UserRole["streamer"]).Where(`users.enable`, 1).Preload(`LiveStreams`, func(db *gorm.DB) *gorm.DB {
			return db.Scopes(StreamsOnline(includeUpcoming))
		}).Preload(`LiveStreams.Match`)
	}
}

func StreamerWithGallery(db *gorm.DB) *gorm.DB {
	return db.Where(`users.role`, consts.UserRole["streamer"]).Where(`users.enable`, 1).Preload(`StreamerGalleries`)
}

func StreamerDefaultPreloads(db *gorm.DB) *gorm.DB {
	return db.Preload(`UserTags`, func(db *gorm.DB) *gorm.DB {
		return db.Where(`status`, 1).Order(`id`)
	}).Preload(`UserAgoraInfo`)
}

// func CheckIfStreamerHasGame(userId, gameId int64) (isFoundBetReport bool, err error) {
// 	var count int64
// 	err = DB.Table("stream_game_users").
// 		Where("stream_game_id = ?", gameId).
// 		Where("user_id = ?", userId).
// 		Count(&count).Error

// 	if count > 0 {
// 		isFoundBetReport = true
// 	}

// 	return
// }
