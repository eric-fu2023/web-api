package model

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Streamer struct {
	models.StreamerC
	IsLive bool
}

func StreamerWithLiveStream(db *gorm.DB) *gorm.DB {
	return db.Where(`users.status`, 1).Where(`users.role`, 2).Preload(`LiveStream`, func(db *gorm.DB) *gorm.DB {
		return db.Where(`live_streams.status`, 3)
	}).Preload(`LiveStream.Match`)
}

func StreamerIsLive(db *gorm.DB) *gorm.DB {
	return db.Joins(`INNER JOIN live_streams ON live_streams.streamer_id = users.id AND live_streams.status = 3`).Order(`live_streams.sort_factor DESC`)
}

func StreamerWithGallery(db *gorm.DB) *gorm.DB {
	return db.Where(`users.status`, 1).Where(`role`, 2).Preload(`UserGalleries`)
}

func TenMostFollowedStreamers(db *gorm.DB) *gorm.DB {
	return db.Scopes(StreamerWithLiveStream).
		Where(`status = 1`).
		Order(`users.followers DESC`).Limit(10)
}

func (a *Streamer) Get(c *gin.Context, id int64) (r Streamer, err error) {
	q := DB.Where(`role`, 2).
		Preload("StreamerStreams", "live_streams.status = 3 OR live_streams.status = 2", func(db *gorm.DB) *gorm.DB {
			return db.Joins(`INNER JOIN real_matches AS matches ON live_streams.match_id = matches.id`).Order("matches.match_time")
		}).
		Preload("StreamerStreams.Match").
		Preload("StreamerStreams.Match.Competition").
		Preload("StreamerStreams.Match.Home").
		Preload("StreamerStreams.Match.Away").
		Preload("UserGalleries").
		Where(`id`, id)
	err = q.First(&r).Error
	return
}

func (a *Streamer) List(c *gin.Context, isLive bool, page int, limit int) (r []Streamer, err error) {
	q := DB.Where(`role`, 2).
		Preload("StreamerStreams", "live_streams.status = 3", func(db *gorm.DB) *gorm.DB {
			return db.Order("status DESC, view_num DESC")
		}).
		Order(`followers DESC, streams DESC`)
	if isLive {
		q = q.Joins(`INNER JOIN live_streams ON live_streams.streamer_id = users.id AND live_streams.status = 3`).Group(`users.id`)
	}
	err = q.Limit(limit).Offset(limit * (page - 1)).Find(&r).Error
	return
}
