package websocket

import (
	"encoding/json"
	"fmt"
	"web-api/util"
)

type RoomMessage struct {
	Id        string `json:"_id,omitempty"`
	SocketId  string `json:"socket_id"`
	Room      string `json:"room"`
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
	UserId    int64  `json:"user_id"`
	UserType  int64  `json:"user_type"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Type      int64  `json:"type"`
	IsHistory bool   `json:"is_history"`
	VipId     int64  `json:"vip_id"`
}

func (a RoomMessage) Send(conn *Connection) (err error) {
	event := "room"
	if a.SocketId != "" {
		event = "room_socket"
	}
	if msg, err := json.Marshal(a); err == nil {
		if err = conn.Send(fmt.Sprintf(`42["%s", %s]`, event, string(msg))); err != nil {
			util.Log().Error("ws send error", err)
		}
	}
	return
}
