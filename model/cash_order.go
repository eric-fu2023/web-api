package model

import (
	"time"

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
			UserId:              userID,
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

func NewCashOutOrder(userID, CashMethodId, amount, balanceBefore int64, account, remark string, reviewRequired bool, accountName string) CashOrder {
	var orderStatus int64 = 1
	if reviewRequired {
		orderStatus = 4
	}

	return CashOrder{
		models.CashOrderC{
			ID:                   models.GenerateCashOutOrderNo(),
			UserId:               userID,
			CashMethodId:         CashMethodId,
			OrderType:            -1,
			Status:               orderStatus,
			AppliedCashOutAmount: amount,
			BalanceBefore:        balanceBefore,
			Account:              account,
			Remark:               remark,
			RequireReview:        reviewRequired,
			AccountName:          accountName,
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

func (CashOrder) List(userID int64, topupOnly, withdrawOnly bool, startTime, endTime *time.Time, page, limit int) (list []CashOrder, err error) {
	db := DB.Scopes(Paginate(page, limit))
	if startTime != nil {
		db = db.Where("created_at > ?", startTime)
	}
	if endTime != nil {
		db = db.Where("created_at < ?", endTime)
	}
	if topupOnly {
		db = db.Where("order_type > 0")
	}
	if withdrawOnly {
		db = db.Where("order_type < 0")
	}

	err = db.
		Where("user_id", userID).
		Order("id DESC").
		Find(&list).Error
	return
}
