package serializer

import (
	"strconv"

	"web-api/model"
	"web-api/service/backend_for_frontend/game_history_pane"
)

type UserCounters struct {
	Order        string            `json:"order"`
	Transaction  string            `json:"transaction"`
	Notification string            `json:"notification"`
	OrderType    map[string]string `json:"order_type"`
}

func BuildUserCounters(a model.UserCounters, _gameOrderHistoriesByPaneType map[game_history_pane.GamesHistoryPaneType]int64, gameHistoryPaneCountsHideAll bool) (UserCounters, error) {
	orderTypes := make(map[string]string)
	if gameHistoryPaneCountsHideAll {
		delete(_gameOrderHistoriesByPaneType, game_history_pane.GamesPaneAll)
	}
	for paneType, count := range _gameOrderHistoriesByPaneType {
		orderTypes[strconv.Itoa(int(paneType))] = formatCounter(count)
	}
	return UserCounters{
		Order:        formatCounter(a.Order),
		Transaction:  formatCounter(a.Transaction),
		Notification: formatCounter(a.Notification),
		OrderType:    orderTypes,
	}, nil
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
