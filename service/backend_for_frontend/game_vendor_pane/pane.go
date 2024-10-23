package game_vendor_pane

import (
	"context"
	"fmt"
	"time"

	"web-api/model"
	"web-api/service"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"blgit.rfdev.tech/taya/ploutos-object/counter"
)

func ResetUserCounter_Order_GamePane(userId int64, paneType service.GamesPaneType) error {
	columnName := ""
	switch paneType {
	case service.GamesPaneTypeCasino:
		columnName = "order_game_pane_casino"
	case service.GamesPaneTypeSports:
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

func CountBetReports(ctx context.Context, userId int64, fromCreatedTime time.Time, gameVendorIds []int64) (int64, error) {
	ctx = rfcontext.AppendCallDesc(ctx, "CountBetReports")
	db := model.DB
	if db == nil {
		return 0, fmt.Errorf("db is nil")
	}
	count, err := counter.CountTable(db.Debug(), ploutos.BetReport{}, model.ByUserId(userId), model.ByCreatedAtGreaterThan(fromCreatedTime))
	return count, err
}
