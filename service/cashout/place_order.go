package cashout

import (
	"errors"
	"fmt"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const userWithdrawLockKey = "user_withdraw_lock.%d"

type WithdrawOrderService struct {
	WithdrawMethodID  int64
	WithdrawAccountNo string
	WithdrawAmount    int64
}

func (s WithdrawOrderService) Do(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)

	u, _ := c.Get("user")
	user := u.(model.User)

	var cl finpay.FinpayClient
	fmt.Print(user)
	fmt.Print(cl)


	mutex := cache.RedisLockClient.NewMutex()
	// check withdrawable
	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		var userSum model.UserSum
		userSum, err = model.UserSum{}.GetByUserIDWithLockWithDB(user.ID, model.DB)
		if err != nil {
			return
		}
		if userSum.MaxWithdrawable < s.WithdrawAmount {
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
		var vipLevel int64 = 0
		var rule model.CashOutRule
		rule, err = model.CashOutRule{}.Get(vipLevel)

		return
	})
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}
	// check cash method rules
	// check vip
	// check cash out rules
	// if trigger review
	// proceed with withdrawal order or hold and review
	// send withdraw request
	cl.WithdrawV1(c)
	return
}

// manual close
// update to pending
// if success update to success
// else update to failed or rejected
