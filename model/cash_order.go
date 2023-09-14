package model

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CashOrder struct {
	models.CashOrderC
}

func NewCashInOrder(userID, CashMethodId, amount, balanceBefore, wagerChange int64, account string) CashOrder {
	return CashOrder{
		models.CashOrderC{
			ID:                  models.GenerateCashInOrderNo(),
			CashMethodId:        CashMethodId,
			OrderType:           1,
			Status:              1,
			AppliedCashInAmount: amount,
			BalanceBefore:       balanceBefore,
			WagerChange:         wagerChange,
			Account:             account,
			//Notes:, update later
		},
	}
}

func (CashOrder) GetPendingWithLockWithDB(orderID string, tx *gorm.DB) (c CashOrder, err error) {
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id", orderID).
		Where("status = 1").
		First(&c).Error
	return
}
