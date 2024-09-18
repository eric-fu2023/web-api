package cashout

import (
	"web-api/model"
	"web-api/serializer"
	"web-api/util"

	"github.com/gin-gonic/gin"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
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
		Where("status", ploutos.CashOrderStatusPendingApproval).
		First(&cashOrder).Error
	if err != nil {
		return
	}

	_, err = RevertCashOutOrder(c, s.OrderNumber, util.JSON(s), "", 5, model.DB)
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
		Where("status", ploutos.CashOrderStatusPendingApproval).
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

	cashOrder, err = DispatchOrder(c, cashOrder, user, cashOrder.UserAccountBinding, s.Retryable)
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return r, err
	}
	r.Data = "Success"
	return
}
