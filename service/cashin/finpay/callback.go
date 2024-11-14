package cashin_finpay

import (
	"context"
	"errors"
	"log"

	"web-api/model"
	"web-api/service/cashin"
	"web-api/service/promotion/on_cash_orders"
	"web-api/util"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
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

	ctx := rfcontext.AppendCallDesc(rfcontext.Spawn(context.Background()), "(s *FinpayPaymentCallback) Handle")
	cashOrder, err := cashin.CloseCashInOrder(c, ctx, s.MerchantOrderNo, s.Amount, 0, 0, util.JSON(s), model.DB, ploutos.TransactionTypeCashIn)

	if err != nil {
		ctx = rfcontext.AppendError(ctx, err, "CloseCashInOrder")
		log.Println(rfcontext.Fmt(ctx))
		return err
	}

	// if err == nil {
	go func() {
		pErr := on_cash_orders.Handle(ctx, cashOrder, ploutos.TransactionTypeCashIn, on_cash_orders.CashOrderEventTypeClose, on_cash_orders.PaymentGatewayFinpay, on_cash_orders.RequestModeCallback)
		if pErr != nil {
			ctx = rfcontext.AppendError(ctx, pErr, "on_cash_orders.Handle")
			log.Println(rfcontext.Fmt(ctx))
		}
	}()
	return nil
}
