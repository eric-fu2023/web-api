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

	GiftId         int64  `json:"gift_id,omitempty"`
	GiftQuantity   int    `json:"gift_quantity,omitempty"`
	GiftName       string `json:"gift_name,omitempty"`
	TotalGiftPrice int64  `json:"total_gift_price,omitempty"`
	IsAnimated     bool   `json:"is_animated,omitempty"`
}

func (a RoomMessage) Send(conn *Connection) (err error) {
	event := "room_system"
	if a.SocketId != "" {
		event = "room_socket"
	}
	var msg []byte
	if msg, err = json.Marshal(a); err != nil {
		util.Log().Error("json marshal error", err)
		return
	}
	if err = conn.Send(fmt.Sprintf(`42["%s", %s]`, event, string(msg))); err != nil {
		util.Log().Error("ws send error", err)
		return
	}
	return
}
