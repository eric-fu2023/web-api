package cashout_finpay

import (
	"web-api/model"
	"web-api/service/cashout"
	"web-api/util"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FinpayTransferCallback struct {
	finpay.TransferCallbackRequest
}

func (s *FinpayTransferCallback) Handle(c *gin.Context) (err error) {
	// if !s.IsValid() {
	// 	err = errors.New("invalid request")
	// 	return
	// }
	if s.IsSucess() {
		_, err = cashout.CloseCashOutOrder(c, s.MerchantOrderNo, int64(s.Amount), 0, 0, util.JSON(s), "", model.DB)
		if err != nil {
			return
		}
	} else if s.IsFailed() {
		err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
			var order model.CashOrder
			err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id", s.MerchantOrderNo).
				Where("status", models.CashOrderStatusCollectionCouldBeTransferring).
				First(&order).Error
			if err != nil {
				return
			}

			_, err = cashout.RevertCashOutOrder(c, order.ID, util.JSON(s), "refund", models.CashOrderStatusFailed, tx)
			if err != nil {
				return
			}
			return
		})
	}
	if err != nil {
		return
	}
	return
}
