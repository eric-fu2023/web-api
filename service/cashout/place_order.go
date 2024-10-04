package cashout

import (
	"context"
	"errors"
	"fmt"
	"time"

	"web-api/cache"
	"web-api/conf"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
	"github.com/go-redsync/redsync/v4"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

const userWithdrawLockKey = "user_withdraw_lock:%d"

type WithdrawOrderService struct {
	Amount            float64 `form:"amount" json:"amount" binding:"required"`
	AccountBindingId  int64   `form:"account_binding_id" json:"account_binding_id" binding:"required"`
	SecondaryPassword string  `form:"secondary_password" json:"secondary_password" binding:"required"`
}

func (s WithdrawOrderService) CreateOrder(c *gin.Context) (r serializer.Response, err error) {
	brand := c.MustGet(`_brand`).(int)
	i18n := c.MustGet("i18n").(i18n.I18n)
	user := c.MustGet("user").(model.User)

	// convert amount from buck to cent
	amountDecimal := decimal.NewFromFloat(s.Amount).IntPart()
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}
	amount := amountDecimal * 100

	// user authentication
	if err = bcrypt.CompareHashAndPassword([]byte(user.SecondaryPassword), []byte(s.SecondaryPassword)); err != nil {
		return serializer.ParamErr(c, s, i18n.T("secondary_password_mismatch"), nil), err
	}

	// retrieve user binded account to withdraw
	var accountBinding ploutos.UserAccountBinding
	err = model.DB.Where("user_id", user.ID).Where("is_active").Where("id", s.AccountBindingId).First(&accountBinding).Error
	if err != nil {
		return
	}
	method, err := model.CashMethod{}.GetByID(c, accountBinding.CashMethodID, brand)
	if err != nil {
		return
	}
	minAmount := method.MinAmount
	err = processCashOutMethod(method)
	if err != nil {
		return
	}
	firstTopup, err := model.FirstTopup(c, user.ID)
	if err != nil || len(firstTopup.ID) == 0 {
		minAmount = conf.GetCfg().WithdrawMinNoDeposit / 100
	}

	if amount < minAmount || amount > method.MaxAmount {
		err = errors.New("illegal amount")
		r = serializer.Err(c, s, serializer.CodeGeneralError, fmt.Sprintf(i18n.T("min_withdraw_unmet"), float64(minAmount)/100), err)
		return
	}

	var reviewRequired bool = false
	var cashOrder ploutos.CashOrder
	vip, err := model.GetVipWithDefault(c, user.ID)
	if err != nil {
		return
	}
	rule := vip.VipRule
	defer func() {
		go func() {
			userSum, _ := model.GetByUserIDWithLockWithDB(user.ID, model.DB)
			common.SendUserSumSocketMsg(user.ID, userSum.UserSum, "withdraw", float64(cashOrder.AppliedCashOutAmount)/100)
		}()
	}()

	mutex := cache.RedisLockClient.NewMutex(fmt.Sprintf(userWithdrawLockKey, user.ID), redsync.WithExpiry(5*time.Second))
	mutex.Lock()
	// check withdrawable
	err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
		var userSum model.UserSum
		userSum, err = model.GetByUserIDWithLockWithDB(user.ID, tx)
		if err != nil {
			return
		}
		if userSum.RemainingWager != 0 || userSum.Balance < amount {
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
		// var rule model.CashOutRule
		var txns []model.Transaction

		timeFrom := time.Now().Truncate(24 * time.Hour)
		txns, err = model.Transaction{}.ListTxRecord(c, user.ID, &timeFrom, nil)
		if err != nil {
			return
		}
		totalOut, payoutCount := CalTxDetails(txns)
		var msg string
		reviewRequired, msg = vipRuleOK(rule, payoutCount, amount, totalOut)

		cashOrder = model.NewCashOutOrder(user.ID, accountBinding.CashMethodID, amount, userSum.Balance, s.AccountBindingId, msg, reviewRequired, c.ClientIP())
		err = tx.Create(&cashOrder).Error
		if err != nil {
			return
		}
		// make balance changes
		// add tx record
		var newUsersum model.UserSum
		newUsersum, err = model.UpdateDbUserSumAndCreateTransaction(rfcontext.AppendCallDesc(rfcontext.Spawn(context.Background()), "WithdrawOrderService"),
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
	// cashOrder, err = DispatchOrder(c, cashOrder, user, accountBinding)
	// if err != nil {
	// 	r = serializer.EnsureErr(c, err, r)
	// 	return
	// }
	// r.Data = serializer.BuildWithdrawOrder(cashOrder)
	return
}

// manual close
// update to pending
// if success update to success
// else update to failed or rejected

func VerifyCashMethod(c *gin.Context, id, amount int64) (err error) {
	brand := c.MustGet(`_brand`).(int)
	cashM, err := model.CashMethod{}.GetByID(c, id, brand)
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

func vipRuleOK(rule ploutos.VIPRule, payoutCount, amount, totalAmount int64) (bool, string) {
	if amount > rule.WithdrawAmount {
		return true, formatMsg("max single withdraw amount exceeded")
	}
	if payoutCount > rule.WithdrawCount {
		return true, formatMsg("max daily withdraw count exceeded")
	}
	if totalAmount > rule.WithdrawAmountTotal {
		return true, formatMsg("max daily withdraw amount exceeded")
	}
	// pending check
	// if c.PayoutDoubleConfirmationRequired && amount > c.PayoutDoubleConfirmationAmount {
	// 	return true, formatMsg("double confirmation triggered")
	// }
	return false, ""
}

func formatMsg(msg string) string {
	return fmt.Sprintf("VIP withdraw rule: %s \n ", msg)
}
