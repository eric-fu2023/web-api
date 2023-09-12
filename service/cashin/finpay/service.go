package cashin

import (
	"fmt"
	"web-api/model"
	"web-api/serializer"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	"github.com/gin-gonic/gin"
)


type TopUpOrderService struct {
	Amount          int64  `form:"amount" json:"amount"`
	Quantity        int64  `form:"quantity" json:"quantity"`
	Sku             string `form:"sku" json:"sku"`
	PaymentMethodID int64  `form:"payment_method_id" json:"payment_method_id"`
}

func (s *TopUpOrderService) CreateOrder(c *gin.Context) (r serializer.Response, err error) {
	u, _ := c.Get("user")
	user := u.(model.User)

	var cl finpay.FinpayClient
	fmt.Print(user)
	fmt.Print(cl)
	// check kyc
	// create cash order
	// create payment order
	// err handling and return
	return
}
