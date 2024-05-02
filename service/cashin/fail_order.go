package cashin

import (
	"context"
	"web-api/model"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

// idempotent
func MarkOrderFailed(c context.Context, orderNumber, notes, transactionID string) (err error) {
	mm := map[string]any{
		"status": models.CashOrderStatusFailed,
		"notes":  models.EncryptedStr(notes),
	}
	if len(transactionID) > 0 {
		mm["transaction_id"] = transactionID
	}
	err = model.DB.Model(&model.CashOrder{}).Where("id", orderNumber).Where("status = 1").
		Updates(mm).Error
	return
}
