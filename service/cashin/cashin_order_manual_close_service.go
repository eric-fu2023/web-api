package cashin

import (
	"context"
	"log"

	"web-api/model"
	"web-api/serializer"
	"web-api/service/promotion/on_cash_orders"
	"web-api/util"

	"blgit.rfdev.tech/taya/common-function/rfcontext"

	"github.com/gin-gonic/gin"
)

type ManualCloseService struct {
	OrderNumber           string `json:"order_number" form:"order_number" binding:"required"`
	ActualAmount          int64  `json:"actual_amount" form:"actual_amount"`
	BonusAmount           int64  `json:"bonus_amount" form:"bonus_amount"`
	AdditionalWagerChange int64  `json:"additional_wager_change" form:"additional_wager_change"`
	TransactionType       int64  `json:"transaction_type," form:"transaction_type"`
}

func (s ManualCloseService) Do(c *gin.Context) (r serializer.Response, err error) {
	if s.TransactionType == 0 {
		s.TransactionType = 10000
	}

	ctx := rfcontext.AppendCallDesc(rfcontext.Spawn(context.Background()), "(s ManualCloseService) Do")
	closedCashInOrder, err := CloseCashInOrder(c, ctx, s.OrderNumber, s.ActualAmount, s.BonusAmount, s.AdditionalWagerChange, util.JSON(s), model.DB, s.TransactionType)
	ctx = rfcontext.AppendError(ctx, err, "CloseCashInOrder")
	log.Println(rfcontext.Fmt(ctx))
	if err != nil {
		r = serializer.Err(c, s, serializer.CodeGeneralError, "", err)
		return
	}
	go func() {
		pErr := on_cash_orders.Handle(ctx, closedCashInOrder, s.TransactionType, on_cash_orders.CashOrderEventTypeClose, on_cash_orders.PaymentGatewayDefault, on_cash_orders.RequestModeManual)
		if pErr != nil {
			ctx = rfcontext.AppendError(ctx, pErr, "on_cash_orders.Handle")
			log.Println(rfcontext.Fmt(ctx))
		}
	}()

	return
}
