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

func GetStreamDetail(streamId int64) (s Stream, err error) {
	err = DB.
		Preload("Streamer").
		Where("id", streamId).
		First(&s).
		Error

	return 
}

func StreamsOnlineSorted(categoryOrder string, categoryTypeOrder string, includeUpcoming bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		order := `sort_factor DESC, schedule_time DESC`
		if len(categoryOrder) > 0 {
			order = `(stream_category_id in ` + categoryOrder + `) DESC, (stream_category_type_id in ` + categoryTypeOrder + `) DESC, ` + order
		}
		return db.Scopes(StreamsOnline(includeUpcoming)).Preload(`Match`).Joins(`INNER JOIN users ON users.id = live_streams.streamer_id AND users.enable = 1`).Order(order)
	}
}

func ExcludeStreamSource(streamerSources []int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(streamerSources) > 0 {
			db = db.Where(`stream_source NOT IN ?`, streamerSources)
		}
		return db
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

func StreamsOnline(includeUpcoming bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		statuses := []int64{2}
		if includeUpcoming {
			statuses = append(statuses, 1)
			db = db.Joins(`LEFT JOIN (SELECT streamer_id, MAX(status) AS status, MIN(schedule_time) AS schedule_time FROM live_streams WHERE status in ? GROUP BY streamer_id) AS upcoming_streams ON upcoming_streams.streamer_id = live_streams.streamer_id AND upcoming_streams.schedule_time = live_streams.schedule_time AND upcoming_streams.status = live_streams.status`, statuses).
				Where(`live_streams.status = 2 OR (live_streams.status = 1 AND upcoming_streams.streamer_id IS NOT NULL)`)
		}
		return db.Where(`live_streams.status`, statuses)
	}
}

func StreamsByFbMatchIdSportId(matchId int64, sportId int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins(`JOIN matches ON live_streams.match_id = matches.id`).Where(`matches.match_id`, matchId).Where(`matches.sport_id`, sportId).
			Where(`live_streams.status`, 2).Order(`live_streams.sort_factor DESC, live_streams.schedule_time`)
	}
}

func StreamsABStreamSource(isA bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if isA {
			return db.Where(`live_streams.stream_source = ?`, 999)
		} else {
			return db.Where(`live_streams.stream_source != ?`, 999)
		}
	}
}
