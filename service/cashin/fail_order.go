package cashin

import (
	"context"
	"web-api/model"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

// idempotent
func MarkOrderFailed(c context.Context, orderNumber, notes string) (err error) {
	err = model.DB.Model(&model.CashOrder{}).Where("id", orderNumber).Updates(map[string]any{
		"status": models.CashOrderStatusFailed,
		"notes":  notes,
	}).Error
	return
}
