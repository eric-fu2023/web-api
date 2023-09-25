package serializer

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type UserNotification struct {
	Text   string `json:"text"`
	Ts     int64  `json:"ts"`
	IsRead bool   `json:"is_read"`
}

func BuildUserNotification(c *gin.Context, a ploutos.UserNotification) (b UserNotification) {
	b = UserNotification{
		Text:   a.Text,
		IsRead: a.IsRead,
	}
	if !a.CreatedAt.IsZero() {
		b.Ts = a.CreatedAt.Unix()
	}
	return
}
