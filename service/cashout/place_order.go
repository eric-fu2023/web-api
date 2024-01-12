package cashout

import (
	"errors"
	"fmt"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"

	"gorm.io/plugin/dbresolver"

	"github.com/gin-gonic/gin"
	"github.com/go-redsync/redsync/v4"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const userWithdrawLockKey = "user_withdraw_lock:%d"

type WithdrawOrderService struct {
	Amount            float64 `form:"amount" json:"amount" binding:"required"`
	AccountBindingID  int64   `form:"account_binding_id" json:"account_binding_id" binding:"required"`
	SecondaryPassword string  `form:"secondary_password" json:"secondary_password" binding:"required"`
}

func (s WithdrawOrderService) Do(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	amountDecimal := decimal.NewFromFloat(s.Amount).IntPart()
	amount := amountDecimal * 100
	user := c.MustGet("user").(model.User)

	if err = bcrypt.CompareHashAndPassword([]byte(user.SecondaryPassword), []byte(s.SecondaryPassword)); err != nil {
		return serializer.ParamErr(c, s, i18n.T("secondary_password_mismatch"), nil), err
	}

	if amount < 0 {
		err = errors.New("illegal amount")
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var accountBinding model.UserAccountBinding
	err = model.DB.Where("user_id", user.ID).Where("is_active").Where("id", s.AccountBindingID).First(&accountBinding).Error
	if err != nil {
		return
	}

	method, err := model.CashMethod{}.GetByID(c, accountBinding.CashMethodID)
	if err != nil {
		return
	}
	err = processCashOutMethod(method)
	if err != nil {
		return
	}

	var reviewRequired bool = false
	var cashOrder model.CashOrder
	mutex := cache.RedisLockClient.NewMutex(fmt.Sprintf(userWithdrawLockKey, user.ID), redsync.WithExpiry(5*time.Second))
	mutex.Lock()
	// check withdrawable
	err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
		var userSum model.UserSum
		userSum, err = model.UserSum{}.GetByUserIDWithLockWithDB(user.ID, tx)
		if err != nil {
			return
		}
		if userSum.MaxWithdrawable < amount || userSum.Balance < amount {
			err = errors.New("withdraw exceeded")
			r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("err_insufficient_withdrawable"), err)
			return
		}
		// var cashMethod model.CashMethod
		// cashMethod, err = model.CashMethod{}.GetByID(c, s.MethodID)
		// if err != nil {
		// 	return
		// }
		// // TODO: check payment method id, if supported, if valid
		// // TODO: check cash method rules maybe
		// fmt.Println(cashMethod, "place holder")

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
		reviewRequired, msg = rule.OK(amount, payoutCount+1, totalOut+amount, user.GetTagIDList())

		cashOrder = model.NewCashOutOrder(user.ID, accountBinding.CashMethodID, amount, userSum.Balance, accountBinding.AccountNumber, msg, reviewRequired, accountBinding.AccountName, c.ClientIP())
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
		notifyBackendWithdraw(cashOrder.ID)
		return
	}
	cashOrder, err = DispatchOrder(c, cashOrder, accountBinding.CashMethodID)
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

func VerifyCashMethod(c *gin.Context, id, amount int64) (err error) {
	cashM, err := model.CashMethod{}.GetByID(c, id)
	if err != nil {
		return err
	}
	if amount > cashM.MaxOneTimePayout || amount < cashM.MinOneTimePayout {
		return ErrCashMethodNotAvailable
	}

	orderList := []model.CashOrder{}
	err = model.DB.Where("cash_method_id", id).Where("order_type", -1).Where("created_at > ?").Find(&orderList).Error
	if err != nil {
		return err
	}

	total := util.Reduce(orderList, func(amount int64, order model.CashOrder) int64 {
		return amount + order.AppliedCashOutAmount
	}, amount)
	if total > cashM.DailyMaxPayout {
		return ErrCashMethodNotAvailable
	}

	return nil
}

var (
	ErrCashMethodNotAvailable = errors.New("cash method limit exceeded")
)
