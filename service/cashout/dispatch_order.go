package cashout

import (
	"errors"
	"web-api/model"
	"web-api/util"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

func DispatchOrder(c *gin.Context, cashOrder model.CashOrder, methodID int64) (updatedCashOrder model.CashOrder, err error) {
	brand := c.MustGet(`_brand`).(int)

	updatedCashOrder = cashOrder
	method, err := model.CashMethod{}.GetByID(c, methodID, brand)
	if err != nil {
		return
	}
	err = processCashOutMethod(method)
	if err != nil {
		return
	}
	switch method.Gateway {
	case "finpay":
		config := method.GetFinpayConfig()
		var data finpay.TransferOrderResponse
		switch config.Type{
		case "TRC20":
			data, err = finpay.FinpayClient{}.DefaultTRC20CashOutV1(c, updatedCashOrder.AppliedCashOutAmount, updatedCashOrder.ID, updatedCashOrder.Account, updatedCashOrder.AccountName)
		default:
			data, err = finpay.FinpayClient{}.DefaultCashOutV1(c, updatedCashOrder.AppliedCashOutAmount, updatedCashOrder.ID, updatedCashOrder.Account, updatedCashOrder.AccountName, config.IfCode, config.BankName, config.Type)
		}
		if data.IsSuccess() {
			updatedCashOrder.Status = models.CashOrderStatusTransferring
			updatedCashOrder.TransactionId = &data.TransferOrderNo
			updatedCashOrder.Notes = util.JSON(data)
			err = model.DB.Debug().WithContext(c).Updates(&updatedCashOrder).Error
			if err != nil {
				return
			}
		} else if data.IsFailed() {
			updatedCashOrder, err = RevertCashOutOrder(c, updatedCashOrder.ID, util.JSON(data), "Request Failed", models.CashOrderStatusFailed, model.DB)
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

func processCashOutMethod(m model.CashMethod) (err error) {
	if m.MethodType > 0 {
		return errors.New("cash method not permitted")
	}
	return nil
}
