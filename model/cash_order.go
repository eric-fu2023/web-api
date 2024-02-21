package model

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	models "blgit.rfdev.tech/taya/ploutos-object"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CashOrder struct {
	ploutos.CashOrder
}

func NewCashInOrder(userID, CashMethodId, amount, balanceBefore, wagerChange int64, account, ip string) CashOrder {
	return CashOrder{
		ploutos.CashOrder{
			ID:                  ploutos.GenerateCashInOrderNo(),
			UserId:              userID,
			CashMethodId:        CashMethodId,
			OrderType:           1,
			Status:              ploutos.CashOrderStatusPending,
			AppliedCashInAmount: amount,
			BalanceBefore:       balanceBefore,
			WagerChange:         wagerChange,
			Account:             account,
			//Notes:, update later
			Ip: ip,
		},
	}
}

func NewCashOutOrder(userID, CashMethodId, amount, balanceBefore int64, account, remark string, reviewRequired bool, accountName, ip string, bankInfo models.BankInfo) CashOrder {
	var orderStatus int64 = 1
	var approveStatus int64
	var reviewStatus int64
	if reviewRequired {
		orderStatus = 4
		approveStatus = 1
		reviewStatus = 1
	}
	raw, _ := json.Marshal(bankInfo)

	return CashOrder{
		ploutos.CashOrder{
			ID:                   ploutos.GenerateCashOutOrderNo(),
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
			ApproveStatus:        approveStatus,
			ReviewStatus:         reviewStatus,
			//Notes:, update later
			Ip:       ip,
			BankInfo: json.RawMessage(raw),
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

func (CashOrder) List(userID int64, topupOnly, withdrawOnly bool, startTime, endTime *time.Time, page, limit int, statusF string) (list []CashOrder, err error) {
	db := DB.Debug().Scopes(Paginate(page, limit))
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
	switch statusF {
	case "success":
		db = db.Where("status in ?", []int64{models.CashOrderStatusSuccess})
	case "failed":
		db = db.Where("status in ?", []int64{models.CashOrderStatusCancelled, models.CashOrderStatusRejected, models.CashOrderStatusFailed})
	case "pending":
		db = db.Where("status in ?", []int64{models.CashOrderStatusPending, models.CashOrderStatusPendingApproval, models.CashOrderStatusTransferring})
	}

	err = db.
		Where("user_id", userID).
		Order("created_at DESC").
		Order("id DESC").
		Find(&list).Error
	return
}

func (CashOrder) IsFirstTime(c context.Context, userID int64) (bool, error) {
	var firstTime bool = false
	err := DB.WithContext(c).Where("user_id", userID).Where("order_type > 0").Where("status", ploutos.CashOrderStatusSuccess).First(&CashOrder{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			firstTime = true
		} else {
			return firstTime, err
		}
	}
	return firstTime, nil
}

func FirstTopup(c context.Context, userID int64) (CashOrder, error) {
	var order CashOrder
	err := DB.WithContext(c).Where("user_id", userID).Where("order_type", models.CashOrderTypeCashIn).
		Where("status", ploutos.CashOrderStatusSuccess).Order("created_at asc").First(&order).Error
	return order, err
}

func ScopedTopupExceptAllTimeFirst(c context.Context, userID int64, start, end time.Time) (list []CashOrder, err error) {
	err = DB.WithContext(c).Where("user_id", userID).Where("order_type", models.CashOrderTypeCashIn).
		Where("status", ploutos.CashOrderStatusSuccess).
		Where("created_at > ?", start).Where("created_at < ?", end).
		Where("id != (?)", DB.WithContext(c).Model(&CashOrder{}).Select("id").Where("user_id", userID).Where("order_type", models.CashOrderTypeCashIn).
			Where("status", ploutos.CashOrderStatusSuccess).Order("created_at asc").Limit(1)).
		Order("created_at asc").Find(&list).Error
	return
}
