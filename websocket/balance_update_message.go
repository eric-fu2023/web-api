package websocket

import (
	"encoding/json"
	"fmt"
	"web-api/util"
)

type BalanceUpdateMessage struct {
	Room            string  `json:"room"`
	Event           string  `json:"event"`
	Cause           string  `json:"cause"`
	Balance         float64 `json:"balance"`
	RemainingWager  float64 `json:"wagering_requirement"`
	MaxWithdrawable float64 `json:"withdrawable"`
	Amount          float64 `json:"amount"`
}

func (a BalanceUpdateMessage) Send(conn *Connection) (err error) {
	if msg, err := json.Marshal(a); err == nil {
		if err = conn.Send(fmt.Sprintf(`42["room_door", %s]`, string(msg))); err != nil {
			util.Log().Error("ws send error", err)
		}
	}
	return
}
