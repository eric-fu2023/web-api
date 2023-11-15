package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type UserNotification struct {
	ID     int64  `json:"id"`
	Text   string `json:"text"`
	Ts     int64  `json:"ts"`
	IsRead bool   `json:"is_read"`
}

func BuildUserNotification(c *gin.Context, a ploutos.UserNotification) (b UserNotification) {
	b = UserNotification{
		ID:     a.ID,
		Text:   a.Text,
		IsRead: a.IsRead,
	}
	if !a.CreatedAt.IsZero() {
		b.Ts = a.CreatedAt.Unix()
	}
	return
}
