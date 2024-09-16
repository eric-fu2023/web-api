package cashout_finpay

import (
	"errors"
	"web-api/model"
	"web-api/service/cashout"
	"web-api/service/promotion/on_cash_orders"
	"web-api/util"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type FinpayTransferCallback struct {
	finpay.TransferCallbackRequest
}

func (s *FinpayTransferCallback) Handle(c *gin.Context) (err error) {
	if !s.IsValid() {
		err = errors.New("invalid request")
		return
	}
	defer model.CashOrder{}.MarkCallbackAt(c, s.MerchantOrderNo, model.DB)

	if s.IsSucess() {
		cashOrder, _err := cashout.CloseCashOutOrder(c, s.MerchantOrderNo, int64(s.Amount), 0, 0, util.JSON(s), "", true, model.DB)
		if err == nil {
			go func() {
				pErr := on_cash_orders.Handle(c, cashOrder, models.TransactionTypeCashOut, on_cash_orders.CashOrderEventTypeClose, on_cash_orders.PaymentGatewayFinPay, on_cash_orders.RequestModeCallback)
				if pErr != nil {
					util.GetLoggerEntry(c).Error("error on promotion handling", pErr)
				}
			}()
		} else {
			err = _err
		}

	} else if s.IsFailed() {
		_, err = cashout.RevertCashOutOrder(c, s.MerchantOrderNo, util.JSON(s), "refund", models.CashOrderStatusFailed, model.DB)
	}
	if err != nil {
		return
	}
	return
}
