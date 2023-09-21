package cashout

import (
	"web-api/model"
	"web-api/serializer"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CloseCashOutOrderService struct {
	OrderNumber  string `form:"order_number" json:"order_number" binding:"required"`
	ActualAmount int64  `form:"actual_amount" json:"actual_amount" binding:"required"`
	BonusAmount  int64  `form:"bonus_amount" json:"bonus_amount"`
	WagerChange  int64  `form:"wager_change" json:"wager_change"`
	// Notes        string `form:"notes" json:"notes" binding:"notes"`
	// Account      string `form:"account" json:"account" binding:"account"`
	Remark string `form:"remark" json:"remark" binding:"remark"`
}

func (s CloseCashOutOrderService) Do(c *gin.Context) (r serializer.Response, err error) {
	_, err = CloseCashOutOrder(c, s.OrderNumber, s.ActualAmount, s.BonusAmount, s.WagerChange, serializer.JSON(s), "", "", model.DB)
	if err != nil {
		r = serializer.GeneralErr(c, err)
		return
	}
	r.Data = "Success"
	return
}

func CloseCashOutOrder(c *gin.Context, orderNumber string, actualAmount, bonusAmount, additionalWagerChange int64, notes, account, remark string, txDB *gorm.DB) (updatedCashOrder model.CashOrder, err error) {
	var newCashOrderState model.CashOrder
	err = txDB.Debug().WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
		err = txDB.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id", orderNumber).First(&newCashOrderState).Error
		if err != nil {
			return
		}
		newCashOrderState.ActualCashOutAmount = actualAmount
		newCashOrderState.BonusCashOutAmount = bonusAmount
		newCashOrderState.EffectiveCashOutAmount = newCashOrderState.AppliedCashOutAmount + bonusAmount
		newCashOrderState.Notes = notes
		newCashOrderState.WagerChange += additionalWagerChange
		newCashOrderState.Account = account
		newCashOrderState.Remark = remark
		newCashOrderState.Status = 2
		// update cash order
		err = tx.Where("id", orderNumber).Updates(newCashOrderState).Error
		if err != nil {
			return
		}
		updatedCashOrder = newCashOrderState
		return
	})

	return
}
