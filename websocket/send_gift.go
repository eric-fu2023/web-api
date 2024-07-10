package websocket

import (
	"encoding/json"
	"fmt"
	"web-api/util"
)

type GiftMessage struct {
	Id           string `json:"_id,omitempty"`
	SocketId     string `json:"socket_id"`
	Room         string `json:"room"`
	Timestamp    int64  `json:"timestamp"`
	Message      string `json:"message"`
	UserId       int64  `json:"user_id"`
	UserType     int64  `json:"user_type"`
	Nickname     string `json:"nickname"`
	Avatar       string `json:"avatar"`
	Type         int64  `json:"type"`
	IsHistory    bool   `json:"is_history"`
	VipId        int64  `json:"vip_id"`
	GiftId       int64  `json:"gift_id"`
	GiftQuantity int    `json:"gift_quantity"`
	GiftName     string `json:"gift_name"`
}

func (giftMsg GiftMessage) Send(conn *Connection) (err error) {
	event := "send_gift"
	if giftMsg.SocketId != "" {
		event = "send_gift_socket"
	}
	var msg []byte
	if msg, err = json.Marshal(giftMsg); err != nil {
		util.Log().Error("json marshal send gift error", err)
		return
	}
	if err = conn.Send(fmt.Sprintf(`42["%s", %s]`, event, string(msg))); err != nil {
		util.Log().Error("ws send gift error", err)
		return
	}
	return
}
