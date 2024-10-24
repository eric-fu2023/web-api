package serializer

import (
	"strconv"

	"web-api/model"
	"web-api/service/backend_for_frontend/game_history_pane"

	"github.com/gin-gonic/gin"
)

type UserCounters struct {
	Order              string             `json:"order"`
	Transaction        string             `json:"transaction"`
	Notification       string             `json:"notification"`
	GameOrderHistories GameOrderHistories `json:"game_order_histories"`
}

type GameOrderHistories struct {
	ByPaneType map[string]string `json:"by_pane_type"`
}

func BuildUserCounters(c *gin.Context, a model.UserCounters, _gameOrderHistoriesByPaneType map[game_history_pane.GamesHistoryPaneType]int64) UserCounters {
	gameOrderHistories := make(map[string]string)
	for paneType, count := range _gameOrderHistoriesByPaneType {
		gameOrderHistories[strconv.Itoa(int(paneType))] = formatCounter(count)
	}
	return UserCounters{
		Order:        formatCounter(a.Order),
		Transaction:  formatCounter(a.Transaction),
		Notification: formatCounter(a.Notification),
		GameOrderHistories: GameOrderHistories{
			ByPaneType: gameOrderHistories,
		},
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
