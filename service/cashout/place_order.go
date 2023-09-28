package cashout

import (
	"errors"
	"fmt"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"

	"github.com/gin-gonic/gin"
	"github.com/go-redsync/redsync/v4"
	"gorm.io/gorm"
)

const userWithdrawLockKey = "user_withdraw_lock:%d"

type WithdrawOrderService struct {
	MethodID    int64  `form:"method_id" json:"method_id" binding:"required"`
	AccountNo   string `form:"account_no" json:"account_no" binding:"required"`
	Amount      int64  `form:"amount" json:"amount" binding:"required,min=0"`
	AccountName string `form:"account_name" json:"account_name" binding:"required"`
}

func (s WithdrawOrderService) Do(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	amount := s.Amount * 100

	switch s.MethodID {
	case 3:
	default:
		err = errors.New("unsupported method")
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	u, _ := c.Get("user")
	user := u.(model.User)

	var reviewRequired bool = false
	var cashOrder model.CashOrder
	mutex := cache.RedisLockClient.NewMutex(fmt.Sprintf(userWithdrawLockKey, user.ID), redsync.WithExpiry(5*time.Second))
	mutex.Lock()
	// check withdrawable
	err = model.DB.Debug().WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
		var userSum model.UserSum
		userSum, err = model.UserSum{}.GetByUserIDWithLockWithDB(user.ID, model.DB.Debug().WithContext(c))
		if err != nil {
			return
		}
		if userSum.MaxWithdrawable < amount || userSum.Balance < amount {
			err = errors.New("withdraw exceeded")
			r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("err_insufficient_withdrawable"), err)
			return
		}
		var cashMethod model.CashMethod
		cashMethod, err = model.CashMethod{}.GetByID(c, s.MethodID)
		if err != nil {
			return
		}
		// TODO: check payment method id, if supported, if valid
		// TODO: check cash method rules maybe
		fmt.Println(cashMethod, "place holder")

		// Get vip level from somewhere else
		// check vip
		// check cash out rules
		var vipLevel int64 = 0
		var rule model.CashOutRule
		var txns []model.Transaction
		rule, err = model.CashOutRule{}.Get(vipLevel)
		if err != nil {
			return
		}

		timeFrom := time.Now().Truncate(24 * time.Hour)
		txns, err = model.Transaction{}.ListTxRecord(c, user.ID, &timeFrom, nil)
		if err != nil {
			return
		}
		totalOut, payoutCount := CalTxDetails(txns)
		var msg string
		reviewRequired, msg = rule.OK(amount, payoutCount+1, totalOut+amount, nil)

		cashOrder = model.NewCashOutOrder(user.ID, s.MethodID, amount, userSum.Balance, s.AccountNo, msg, reviewRequired, s.AccountName)
		err = tx.Create(&cashOrder).Error
		if err != nil {
			return
		}
		// make balance changes
		// add tx record
		var newUsersum model.UserSum
		newUsersum, err = model.UserSum{}.UpdateUserSumWithDB(
			tx,
			user.ID,
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
	r.Data = serializer.BuildWithdrawOrder(cashOrder)
	if reviewRequired {
		return
	}
	cashOrder, err = DispatchOrder(c, cashOrder, s.MethodID)
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}
	r.Data = serializer.BuildWithdrawOrder(cashOrder)
	return
}

// manual close
// update to pending
// if success update to success
// else update to failed or rejected
