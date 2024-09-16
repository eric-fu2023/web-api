package cashout_finpay

import (
	"web-api/model"
	"web-api/serializer"
	"web-api/service/cashout"
	"web-api/service/promotion/on_cash_orders"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type CashOutOrderService struct {
	OrderNumber string `form:"order_number" json:"order_number" binding:"required"`
	Retryable   bool   `form:"retryable" json:"retryable"`
	// ActualAmount int64  `form:"actual_amount" json:"actual_amount" binding:"required"`
	// BonusAmount  int64  `form:"bonus_amount" json:"bonus_amount"`
	// WagerChange  int64  `form:"wager_change" json:"wager_change"`
	// // Notes        string `form:"notes" json:"notes" binding:"notes"`
	// // Account      string `form:"account" json:"account" binding:"account"`
	// Remark string `form:"remark" json:"remark"`
}

func (s CashOutOrderService) Reject(c *gin.Context) (r serializer.Response, err error) {

	var cashOrder model.CashOrder
	err = model.DB.Where("id", s.OrderNumber).
		Where("status", models.CashOrderStatusPendingApproval).
		First(&cashOrder).Error
	if err != nil {
		return
	}

	_, err = cashout.RevertCashOutOrder(c, s.OrderNumber, util.JSON(s), "", 5, model.DB)
	if err != nil {
		r = serializer.GeneralErr(c, err)
		return
	}
	r.Data = "Success"
	return
}

func (s CashOutOrderService) Approve(c *gin.Context) (r serializer.Response, err error) {
	var cashOrder model.CashOrder
	err = model.DB.Debug().Preload("UserAccountBinding").Where("id", s.OrderNumber).
		Where("review_status", 2).
		Where("status", models.CashOrderStatusPendingApproval).
		First(&cashOrder).Error
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}
	var user model.User
	if err = model.DB.Where(`id`, cashOrder.UserId).First(&user).Error; err != nil {
		util.GetLoggerEntry(c).Error("get user error")
		return
	}

	cashOrder, err = cashout.DispatchOrder(c, cashOrder, user, cashOrder.UserAccountBinding, s.Retryable)
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return r, err
	}
	r.Data = "Success"
	return
}

type ManualCloseOrderService struct {
	OrderNumber string `form:"order_number" json:"order_number" binding:"required"`
}

func (s ManualCloseOrderService) Do(c *gin.Context) (r serializer.Response, err error) {
	var cashOrder model.CashOrder
	err = model.DB.Debug().Preload("UserAccountBinding").Where("id", s.OrderNumber).
		Where("review_status", 2).
		Where("status", models.CashOrderStatusPendingApproval).
		First(&cashOrder).Error
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}

	tx := model.DB.Begin()
	_, err = cashout.CloseCashOutOrder(c, s.OrderNumber, cashOrder.AppliedCashOutAmount, 0, 0, util.JSON(s), "", false, tx)
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

	go func() {
		pErr := on_cash_orders.Handle(c, cashOrder, models.TransactionTypeCashOut, on_cash_orders.CashOrderEventTypeClose, on_cash_orders.PaymentGatewayForay, on_cash_orders.RequestModeManual)
		if pErr != nil {
			util.GetLoggerEntry(c).Error("error on promotion handling", pErr)
		}
	}()

	return
}

// type CloseCashOutOrderService struct {
// 	OrderNumber  string `form:"order_number" json:"order_number" binding:"required"`
// 	ActualAmount int64  `form:"actual_amount" json:"actual_amount" binding:"required"`
// 	BonusAmount  int64  `form:"bonus_amount" json:"bonus_amount"`
// 	WagerChange  int64  `form:"wager_change" json:"wager_change"`
// 	// Notes        string `form:"notes" json:"notes" binding:"notes"`
// 	// Account      string `form:"account" json:"account" binding:"account"`
// }

// func (s CloseCashOutOrderService) Do(c *gin.Context) (r serializer.Response, err error) {
// 	_, err = cashout.CloseCashOutOrder(c, s.OrderNumber, s.ActualAmount, s.BonusAmount, s.WagerChange, util.JSON(s), "", model.DB)
// 	if err != nil {
// 		r = serializer.GeneralErr(c, err)
// 		return
// 	}
// 	r.Data = "Success"
// 	return
// }

// func (s CancelCashOutOrderService) Cancel(c *gin.Context) (r serializer.Response, err error) {
// 	_, err = cashout.RevertCashOutOrder(c, s.OrderNumber, util.JSON(s), s.Remark, 3, model.DB)
// 	if err != nil {
// 		r = serializer.GeneralErr(c, err)
// 		return
// 	}
// 	r.Data = "Success"
// 	return
// }
