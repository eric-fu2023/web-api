package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
	"web-api/conf/consts"
)

type Streamer struct {
	ploutos.User
	IsLive bool
}

func StreamerWithLiveStream(db *gorm.DB) *gorm.DB {
	return db.Where(`users.role`, consts.UserRole["streamer"]).Where(`users.enable`, 1).Preload(`LiveStreams`, func(db *gorm.DB) *gorm.DB {
		return db.Scopes(StreamsOnline)
	}).Preload(`LiveStreams.Match`)
}

func StreamerWithGallery(db *gorm.DB) *gorm.DB {
	return db.Where(`users.role`, consts.UserRole["streamer"]).Where(`users.enable`, 1).Preload(`StreamerGalleries`)
}
