package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type StreamAnnouncement struct {
	ploutos.StreamAnnouncement

	LiveStream Stream `gorm:"foreignKey:StreamId;references:ID"`
}

func GetStreamerAnnouncement(streamerId int64) (list []StreamAnnouncement, err error) {
	err = DB.
		Joins("JOIN live_streams ON live_streams.id = stream_announcements.stream_id").
		Scopes(StreamsOnline(false), FilterByStreamer(streamerId)).
		Limit(2).
		Order("created_at asc").
		Find(&list).
		Error

	return 
}
 
