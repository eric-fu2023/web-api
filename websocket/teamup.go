package websocket

import (
	"encoding/json"
	"fmt"
	"web-api/util"
)

type TeamupGameNotificationMessage struct {
	UserId       int64  `json:"user_id"`
	Amount       int64  `json:"amount"`
	Icon         string `json:"icon"`
	Event        string `json:"event"`
	ProviderName string `json:"provider_name"`
	EndTime      int64  `json:"end_time"`
	TeamupId     int64  `json:"teamup_id"`
	TeamupType   int64  `json:"teamup_type"`
}

func (msg TeamupGameNotificationMessage) Send(conn *Connection) (err error) {
	if msg, err := json.Marshal(msg); err == nil {
		if err = conn.Send(fmt.Sprintf(`42["common", %s]`, string(msg))); err != nil {
			util.Log().Error("ws send error", err)
		}
	}

	return
}
