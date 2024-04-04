package cashin

import (
	"context"
	"web-api/model"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

// idempotent
func MarkOrderFailed(c context.Context, orderNumber, notes, transactionID string) (err error) {
	err = model.DB.Model(&model.CashOrder{}).Where("id", orderNumber).Where("status = 1").
		Updates(map[string]any{
			"status":         models.CashOrderStatusFailed,
			"notes":          models.EncryptedStr(notes),
			"transaction_id": transactionID,
		}).Error
	return
}
