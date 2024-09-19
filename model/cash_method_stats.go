package model

import (
	"fmt"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

// success/failed/gateway_failed
func IncrementCashMethodStats(stats ploutos.CashMethodStats, result string) error {
	field := ""
	switch result {
	case "success":
		field = "success"
	case "failed":
		field = "failed"
	case "gateway_failed":
		field = "gateway_failed"
	}
	err := DB.Debug().Exec(fmt.Sprintf("UPDATE cash_method_stats SET called = called + 1, %s = %s + 1 WHERE id = ?", field, field), stats.ID).Error
	return err
}
