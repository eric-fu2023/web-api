package cashout

import (
	"errors"
	"fmt"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"

	"github.com/gin-gonic/gin"
	"github.com/go-redsync/redsync/v4"
	"gorm.io/gorm"
)

type CustomOrderService struct {
	model.CashOrder
}

func (cashOrder CustomOrderService) Handle(c *gin.Context) (r serializer.Response, err error) {
	amount := cashOrder.AppliedCashOutAmount
	mutex := cache.RedisLockClient.NewMutex(fmt.Sprintf(userWithdrawLockKey, cashOrder.UserId), redsync.WithExpiry(5*time.Second))
	mutex.Lock()
	// check withdrawable
	err = model.DB.Debug().WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
		var userSum model.UserSum
		userSum, err = model.UserSum{}.GetByUserIDWithLockWithDB(cashOrder.UserId, model.DB.Debug().WithContext(c))
		if err != nil {
			return
		}
		if userSum.MaxWithdrawable < amount || userSum.Balance < amount {
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
		newUsersum, err = model.UserSum{}.UpdateUserSumWithDB(
			tx,
			cashOrder.UserId,
			-amount,
			0,
			-amount, 10001, cashOrder.ID)
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
