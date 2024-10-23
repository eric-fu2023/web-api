package game_history_pane

import (
	"fmt"
	"time"

	"web-api/model"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

var GamesPaneLastSeenTypeToUserCounterColumn = map[GamesHistoryPaneType]string{
	GamesPaneTypeCasino: "game_history_casino_last_seen",
	GamesPaneTypeSports: "game_history_sports_last_seen",
}

const UserCounterColumnTransactionLastSeen = `transaction_last_seen`
const UserCounterColumnNotificationLastSeen = `notification_last_seen`

func AdvanceUserCounter_Order_GamePane_LastSeen(userId int64, paneType GamesHistoryPaneType, datetime time.Time) error {
	columnName, ok := GamesPaneLastSeenTypeToUserCounterColumn[paneType]
	if !ok {
		return fmt.Errorf("column name for pane type %d not exist", paneType)
	}

	return _userCounterAdvanceLastSeen(userId, datetime, columnName)
}

func AdvanceTransactionLastSeen(userId int64, datetime time.Time) error {
	return _userCounterAdvanceLastSeen(userId, datetime, UserCounterColumnTransactionLastSeen)
}

func AdvanceNotificationLastSeen(userId int64, datetime time.Time) error {
	return _userCounterAdvanceLastSeen(userId, datetime, UserCounterColumnNotificationLastSeen)
}

func _userCounterAdvanceLastSeen(userId int64, datetime time.Time, columnName string) error {
	err := model.DB.Debug().Model(ploutos.UserCounter{}).Scopes(model.ByUserId(userId)).Update(columnName, datetime).Error
	return err
}

func ResetUserCounter_OrderCount(userId int64) {
	model.DB.Model(ploutos.UserCounter{}).Scopes(model.ByUserId(userId)).Updates(map[string]interface{}{"order_count": 0, "order_last_seen": time.Now()})
}
