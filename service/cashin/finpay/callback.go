package cashin_finpay

import (
	"errors"

	"web-api/model"
	"web-api/service/cashin"
	"web-api/service/promotion/on_cash_orders"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	"github.com/gin-gonic/gin"
)

type FinpayPaymentCallback struct {
	finpay.PaymentOrderCallBackReq
}

func (s *FinpayPaymentCallback) Handle(c *gin.Context) error {
	if !s.IsValid() {
		return errors.New("invalid response")
	}
	defer model.CashOrder{}.MarkCallbackAt(c, s.MerchantOrderNo, model.DB)

	if !s.IsSuccess() {
		return cashin.MarkOrderFailed(c, s.MerchantOrderNo, util.JSON(s), s.PaymentOrderNo)
	}
	// check api response
	// lock cash order
	// update cash order
	// {
	// update user_sum
	// create transaction history
	// }
	cashOrder, err := cashin.CloseCashInOrder(c, s.MerchantOrderNo, s.Amount, 0, 0, util.JSON(s), model.DB, ploutos.TransactionTypeCashIn)
	if err != nil {
		return err
	}

	// if err == nil {
	go func() {
		pErr := on_cash_orders.Handle(c.Copy(), cashOrder, ploutos.TransactionTypeCashIn, on_cash_orders.CashOrderEventTypeClose, on_cash_orders.PaymentGatewayFinpay, on_cash_orders.RequestModeCallback)
		if pErr != nil {
			util.GetLoggerEntry(c).Error("error on promotion handling", pErr)
		}
	}()
	return nil
}
