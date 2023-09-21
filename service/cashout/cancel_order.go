package cashout

import (
	"errors"
	"web-api/model"
	"web-api/serializer"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CancelCashOutOrderService struct {
	OrderNumber  string `form:"order_number" json:"order_number" binding:"required"`
	ActualAmount int64  `form:"actual_amount" json:"actual_amount" binding:"required"`
	BonusAmount  int64  `form:"bonus_amount" json:"bonus_amount"`
	WagerChange  int64  `form:"wager_change" json:"wager_change"`
	// Notes        string `form:"notes" json:"notes" binding:"notes"`
	// Account      string `form:"account" json:"account" binding:"account"`
	Remark string `form:"remark" json:"remark"`
}

func (s CancelCashOutOrderService) Reject(c *gin.Context) (r serializer.Response, err error) {
	_, err = RevertCashOutOrder(c, s.OrderNumber, serializer.JSON(s), "", s.Remark, 5, model.DB)
	if err != nil {
		r = serializer.GeneralErr(c, err)
		return
	}
	r.Data = "Success"
	return
}

func (s CancelCashOutOrderService) Cancel(c *gin.Context) (r serializer.Response, err error) {
	_, err = RevertCashOutOrder(c, s.OrderNumber, serializer.JSON(s), "", s.Remark, 3, model.DB)
	if err != nil {
		r = serializer.GeneralErr(c, err)
		return
	}
	r.Data = "Success"
	return
}

// newStatus = 3, 5
func RevertCashOutOrder(c *gin.Context, orderNumber string, notes, account, remark string, newStatus int64, txDB *gorm.DB) (updatedCashOrder model.CashOrder, err error) {
	var newCashOrderState model.CashOrder
	switch newStatus {
	case 3, 5:
	default:
		err = errors.New("wrong status")
		return
	}
	err = txDB.Transaction(func(tx *gorm.DB) (err error) {
		err = txDB.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id", orderNumber).First(&newCashOrderState).Error
		if err != nil {
			return
		}

		newCashOrderState.Notes = notes
		newCashOrderState.Account = account
		newCashOrderState.Remark = remark
		newCashOrderState.Status = newStatus
		// update cash order
		err = tx.Where("id", orderNumber).Updates(newCashOrderState).Error
		if err != nil {
			return
		}
		updatedCashOrder = newCashOrderState
		_, err = model.UserSum{}.UpdateUserSumWithDB(
			tx,
			newCashOrderState.UserId,
			newCashOrderState.AppliedCashOutAmount,
			newCashOrderState.WagerChange,
			newCashOrderState.AppliedCashOutAmount,
			10002,
			newCashOrderState.ID)

		return
	})

	return
}
