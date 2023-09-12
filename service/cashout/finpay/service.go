package cashout

import (
	"fmt"
	"web-api/model"
	"web-api/serializer"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	"github.com/gin-gonic/gin"
)

type WithdrawOrderService struct{
	WithdrawMethodID int64
	WithdrawAccountNo string
	WithdrawAmount int64
}

func (s WithdrawOrderService) Do(c *gin.Context) (r serializer.Response, err error) {
	u, _ := c.Get("user")
	user := u.(model.User)

	var cl finpay.FinpayClient
	fmt.Print(user)
	fmt.Print(cl)
	// check withdrawable
	// check cash method rules
	// check vip
	// check cash out rules
	// if trigger review
	// proceed with withdrawal order or hold and review
	cl.WithdrawV1(c)
	return
}