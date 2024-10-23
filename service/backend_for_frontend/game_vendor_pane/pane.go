package game_vendor_pane

import (
	"fmt"
	"time"

	"web-api/model"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

func ResetUserCounter_Order_GamePane(userId int64, paneType GamesPaneType) error {
	columnName := ""
	switch paneType {
	case GamesPaneTypeCasino:
		columnName = "order_game_pane_casino"
	case GamesPaneTypeSports:
		columnName = "order_game_pane_sports"
	default:
		return fmt.Errorf("order_pane")
	}

	err := model.DB.Model(ploutos.UserCounter{}).Scopes(model.ByUserId(userId)).Updates(map[string]interface{}{columnName: 0, "order_last_seen": time.Now()}).Error
	if err != nil {
		return err
	}

	return nil
}
