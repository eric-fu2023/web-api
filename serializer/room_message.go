package serializer

import (
	"web-api/model"
)

type RoomMessage struct {
	Room      string `json:"room"`
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	UserId    int64  `json:"user_id"`
	UserType  int64  `json:"user_type"`
	Type      int64  `json:"type"`
}

func BuildRoomMessage(a model.RoomMessage) (b RoomMessage) {
	b = RoomMessage{
		Room:      a.Room,
		Message:   a.Message,
		Nickname:  a.Nickname,
		Avatar:    a.Avatar,
		UserType:  a.UserType,
		Type:      a.Type,
		UserId:    a.UserId,
		Timestamp: a.Timestamp,
	}
	return
}
