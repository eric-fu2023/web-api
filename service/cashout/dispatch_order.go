package cashout

import (
	"web-api/model"
	"web-api/util"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

func DispatchOrder(c *gin.Context, cashOrder model.CashOrder, methodID int64) (updatedCashOrder model.CashOrder, err error) {
	updatedCashOrder = cashOrder
	switch methodID {
	case 3:
		var data finpay.TransferOrderResponse
		data, err = finpay.FinpayClient{}.DefaultGcashCashOutV1(c, updatedCashOrder.AppliedCashOutAmount, updatedCashOrder.ID, updatedCashOrder.Account, updatedCashOrder.AccountName)
		if data.IsSuccess() {
			updatedCashOrder.Status = models.CashOrderStatusTransferring
			updatedCashOrder.TransactionId = data.TransferId
			updatedCashOrder.Notes = util.JSON(data)
			err = model.DB.Updates(&updatedCashOrder).Error
			if err != nil {
				return
			}
		} else if data.IsFailed() {
			updatedCashOrder, err = RevertCashOutOrder(c, updatedCashOrder.ID, util.JSON(data), "Request Failed", 7, model.DB)
			if err != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}
	return
}
