package model

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

type Streamer struct {
	models.Streamer
	IsLive bool
}

func StreamerWithLiveStream(db *gorm.DB) *gorm.DB {
	return db.Where(`streamers.enable`, 1).Preload(`LiveStreams`, func(db *gorm.DB) *gorm.DB {
		return db.Scopes(StreamsOnline)
	}).Preload(`LiveStreams.Match`)
}

func StreamerIsLive(db *gorm.DB) *gorm.DB {
	return db.Joins(`INNER JOIN live_streams ON live_streams.streamer_id = users.id AND live_streams.status = 3`).Order(`live_streams.sort_factor DESC`)
}

func StreamerWithGallery(db *gorm.DB) *gorm.DB {
	return db.Where(`streamers.enable`, 1).Preload(`StreamerGalleries`)
}

func TenMostFollowedStreamers(db *gorm.DB) *gorm.DB {
	return db.Scopes(StreamerWithLiveStream).
		Where(`status = 1`).
		Order(`users.followers DESC`).Limit(10)
}
