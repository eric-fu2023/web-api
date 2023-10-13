package model

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

type Stream struct {
	models.LiveStream
	Streamer *Streamer `gorm:"references:StreamerId;foreignKey:ID"`
	Match    *Match    `gorm:"references:MatchId;foreignKey:ID"`
}

func StreamsOnlineSorted(categoryOrder string, categoryTypeOrder string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		order := `sort_factor DESC, schedule_time DESC`
		if len(categoryOrder) > 0 {
			order = `(stream_category_id in ` + categoryOrder + `) DESC, (stream_category_type_id in ` + categoryTypeOrder + `) DESC, ` + order
		}
		return db.Scopes(StreamsOnline).Preload(`Match`).Joins(`INNER JOIN users ON users.id = live_streams.streamer_id AND users.enable = 1`).Order(order)
	}
}

func ExcludeStreamers(streamerIds []int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(streamerIds) > 0 {
			db = db.Where(`streamer_id NOT IN ?`, streamerIds)
		}
		return db
	}
}

func StreamsOnline(db *gorm.DB) *gorm.DB {
	return db.Where(`live_streams.status`, 2)
}

func StreamsByMatchId(matchId int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`match_id`, matchId).
			Where(`status`, 2).Order(`sort_factor DESC, schedule_time`)
	}
}
