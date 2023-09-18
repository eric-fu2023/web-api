package cashout

import (
	"errors"
	"fmt"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	"github.com/gin-gonic/gin"
	"github.com/go-redsync/redsync/v4"
	"gorm.io/gorm"
)

const userWithdrawLockKey = "user_withdraw_lock.%d"

type WithdrawOrderService struct {
	WithdrawMethodID  int64
	WithdrawAccountNo string
	WithdrawAmount    int64 ` binding:"required, min=0"`
}

func (s WithdrawOrderService) Do(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)

	u, _ := c.Get("user")
	user := u.(model.User)

	var reviewRequired bool = false
	mutex := cache.RedisLockClient.NewMutex(fmt.Sprintf(userWithdrawLockKey, user.ID), redsync.WithExpiry(5*time.Second))
	mutex.Lock()
	// check withdrawable
	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		var userSum model.UserSum
		userSum, err = model.UserSum{}.GetByUserIDWithLockWithDB(user.ID, model.DB)
		if err != nil {
			return
		}
		if userSum.MaxWithdrawable < s.WithdrawAmount || userSum.Balance < s.WithdrawAmount {
			err = errors.New("withdraw exceeded")
			r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("err_insufficient_withdrawable"), err)
			return
		}
		var cashMethod model.CashMethod
		cashMethod, err = model.CashMethod{}.GetByID(c, s.WithdrawMethodID)
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
		reviewRequired, msg = rule.OK(s.WithdrawAmount, payoutCount+1, totalOut+s.WithdrawAmount, nil)

		cashOrder := model.NewCashOutOrder(user.ID, s.WithdrawMethodID, s.WithdrawAmount, userSum.Balance, s.WithdrawAccountNo, msg, reviewRequired)
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
			-s.WithdrawAmount, 
			0, 
			-s.WithdrawAmount, 10001, cashOrder.ID)
		if err != nil {
			return
		}
		// sanity check
		if newUsersum.Balance != userSum.Balance-s.WithdrawAmount||
		newUsersum.MaxWithdrawable != userSum.MaxWithdrawable-s.WithdrawAmount ||
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
	if reviewRequired {
		return
	}

	finpay.FinpayClient{}.WithdrawV1(c)
	return
}

// manual close
// update to pending
// if success update to success
// else update to failed or rejected
