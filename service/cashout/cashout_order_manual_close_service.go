package cashout

import (
	"web-api/model"
	"web-api/serializer"
	"web-api/service/promotion/on_cash_orders"
	"web-api/util"

	"github.com/gin-gonic/gin"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type ManualCloseOrderService struct {
	OrderNumber string `form:"order_number" json:"order_number" binding:"required"`
}

func (s ManualCloseOrderService) Do(c *gin.Context) (r serializer.Response, err error) {
	var cashOrder model.CashOrder
	err = model.DB.Debug().Preload("UserAccountBinding").Where("id", s.OrderNumber).
		Where("review_status", 2).
		Where("status", ploutos.CashOrderStatusPendingApproval).
		First(&cashOrder).Error
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}

	tx := model.DB.Begin()
	_, err = CloseCashOutOrder(c, s.OrderNumber, cashOrder.AppliedCashOutAmount, 0, 0, util.JSON(s), "", false, tx)
	if err != nil {
		tx.Rollback()
		r = serializer.EnsureErr(c, err, r)
		return
	}
	cashOrder.IsManualCashOut = true

	err = tx.Debug().Select("IsManualCashOut").Updates(&cashOrder).Error
	if err != nil {
		tx.Rollback()
		r = serializer.EnsureErr(c, err, r)
		return
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		r = serializer.EnsureErr(c, err, r)
		return
	}
	pErr := on_cash_orders.Handle(c, cashOrder, ploutos.TransactionTypeCashOut, on_cash_orders.CashOrderEventTypeClose, on_cash_orders.PaymentGatewayFinPay, on_cash_orders.RequestModeManual)
	if pErr != nil {
		util.GetLoggerEntry(c).Error("error on promotion handling", pErr)
	}

	return
}
