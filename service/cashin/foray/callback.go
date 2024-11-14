package cashin_foray

import (
	"context"
	"errors"
	"log"

	"web-api/model"
	"web-api/service/cashin"
	"web-api/service/promotion/on_cash_orders"
	"web-api/util"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	forayDto "blgit.rfdev.tech/taya/payment-service/foray/dto"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type ForayPaymentCallback struct {
	forayDto.DepositOrderCallbackRequest
}

func (s *ForayPaymentCallback) Handle(c *gin.Context) error {
	if !s.IsValid() {
		return errors.New("invalid response")
	}
	defer model.CashOrder{}.MarkCallbackAt(c, s.OrderTranoIn, model.DB)

	if !s.IsSuccess() {
		return cashin.MarkOrderFailed(c, s.OrderTranoIn, util.JSON(s), s.OrderNumber)
	}
	// check api response
	// lock cash order
	// update cash order
	// {
	// update user_sum
	// create transaction history
	// }

	ctx := rfcontext.AppendCallDesc(rfcontext.Spawn(context.Background()), "(s *ForayPaymentCallback) Handle")
	txType := ploutos.TransactionTypeCashIn
	cashOrder, err := cashin.CloseCashInOrder(c, ctx, s.OrderTranoIn, s.OrderAmount, 0, 0, util.JSON(s), model.DB, txType)

	if err != nil {
		ctx = rfcontext.AppendError(ctx, err, "CloseCashInOrder")
		log.Println(rfcontext.Fmt(ctx))
		return err
	}
	// if err == nil {
	go func() {
		pErr := on_cash_orders.Handle(ctx, cashOrder, txType, on_cash_orders.CashOrderEventTypeClose, on_cash_orders.PaymentGatewayForay, on_cash_orders.RequestModeCallback)
		if pErr != nil {
			ctx = rfcontext.AppendError(ctx, pErr, "on_cash_orders.Handle")
			log.Println(rfcontext.Fmt(ctx))
		}
	}()

	return nil
}
