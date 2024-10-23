package serializer

import (
	"strconv"

	"web-api/model"
	"web-api/service/backend_for_frontend/game_history_pane"

	"github.com/gin-gonic/gin"
)

type UserCounters struct {
	Order              string            `json:"order"`
	Transaction        string            `json:"transaction"`
	Notification       string            `json:"notification"`
	GameOrderHistories map[string]string `json:"game_order_histories"`
}

func BuildUserCounters(c *gin.Context, a model.UserCounters, _gameOrderHistories map[game_history_pane.GamesHistoryPaneType]int64) UserCounters {
	gameOrderHistories := make(map[string]string)
	for paneType, count := range _gameOrderHistories {
		gameOrderHistories[strconv.Itoa(int(paneType))] = strconv.Itoa(int(count))
	}
	return UserCounters{
		Order:              formatCounter(a.Order),
		Transaction:        formatCounter(a.Transaction),
		Notification:       formatCounter(a.Notification),
		GameOrderHistories: gameOrderHistories,
	}
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
