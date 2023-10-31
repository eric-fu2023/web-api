package serializer

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"web-api/model"
)

type UserCounters struct {
	Order        string `json:"order"`
	Transaction  string `json:"transaction"`
	Notification string `json:"notification"`
}

func BuildUserCounters(c *gin.Context, a model.UserCounters) (b UserCounters) {
	b = UserCounters{
		Order:        formatCounter(a.Order),
		Transaction:  formatCounter(a.Transaction),
		Notification: formatCounter(a.Notification),
	}
	return
}

func formatCounter(counter int64) string {
	if counter > 99 {
		return "99+"
	} else if counter > 0 {
		return strconv.Itoa(int(counter))
	} else {
		return "0"
	}
}
