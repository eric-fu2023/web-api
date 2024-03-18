package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/json"
)

type PrivateMessage struct {
	ID        int64       `json:"id"`
	Type      int64       `json:"type"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

func BuildPrivateMessage(a ploutos.PrivateMessage) (b PrivateMessage) {
	b = PrivateMessage{
		ID:        a.ID,
		Type:      a.Type,
		Message:   a.Message,
		Timestamp: a.CreatedAt.Unix(),
	}
	if a.Data != nil {
		var t map[string]interface{}
		if e := json.Unmarshal(a.Data, &t); e == nil {
			b.Data = t
		}
	}
	return
}
