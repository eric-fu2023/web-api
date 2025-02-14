package cashout

import (
	"context"
	"errors"
	"fmt"
	"time"

	"web-api/cache"
	"web-api/model"
	"web-api/serializer"

	"blgit.rfdev.tech/taya/common-function/rfcontext"

	"github.com/gin-gonic/gin"
	"github.com/go-redsync/redsync/v4"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type CustomOrderService struct {
	model.CashOrder
	TransactionType int64
}

func (svc CustomOrderService) Handle(c *gin.Context) (r serializer.Response, err error) {
	cashOrder := svc.CashOrder
	if svc.TransactionType == 0 {
		svc.TransactionType = 10001
	}
	amount := cashOrder.AppliedCashOutAmount

	verified := svc.VerifyManualCashOut()
	if !verified {
		err = errors.New("invalid cashout order")
		r = serializer.Err(c, nil, serializer.CodeGeneralError, "err_invalid_order", err)
		return
	}

	mutex := cache.RedisLockClient.NewMutex(fmt.Sprintf(userWithdrawLockKey, cashOrder.UserId), redsync.WithExpiry(5*time.Second))
	mutex.Lock()
	// check withdrawable
	err = model.DB.Clauses(dbresolver.Use("txConn")).WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
		var userSum model.UserSum
		userSum, err = model.GetByUserIDWithLockWithDB(cashOrder.UserId, model.DB.Debug().WithContext(c))
		if err != nil {
			return
		}
		if userSum.Balance < amount {
			err = errors.New("withdraw exceeded")
			r = serializer.Err(c, nil, serializer.CodeGeneralError, "err_insufficient_withdrawable", err)
			return
		}
		// TODO: check payment method id, if supported, if valid
		// TODO: check cash method rules maybe
		// fmt.Println(cashMethod, "place holder")

		// Get vip level from somewhere else
		// check vip
		// check cash out rules

		err = tx.Create(&cashOrder).Error
		if err != nil {
			return
		}
		// make balance changes
		// add tx record
		var newUsersum model.UserSum
		newUsersum, err = model.UpdateDbUserSumAndCreateTransaction(rfcontext.AppendCallDesc(rfcontext.Spawn(context.Background()), "CustomOrderService"),
			tx,
			cashOrder.UserId,
			-amount,
			0,
			-amount, svc.TransactionType, cashOrder.ID)
		if err != nil {
			return
		}
		// sanity check
		if newUsersum.Balance != userSum.Balance-amount ||
			newUsersum.MaxWithdrawable != userSum.MaxWithdrawable-amount ||
			newUsersum.RemainingWager != userSum.RemainingWager {
			err = errors.New("error handling user balance")
			return
		}
		// if trigger review
		// proceed with withdrawal order or hold and review
		// send withdraw request
		return
	})
	mutex.Unlock()

	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}
	return
}

func (s CustomOrderService) VerifyManualCashOut() bool {
	return s.RequireReview && s.ReviewStatus == 1
}
