package cashin_finpay

import (
	"errors"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/cashin"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	"github.com/gin-gonic/gin"
)

type FinpayCallback struct {
	finpay.PaymentOrderCallBackReq
}

func (s *FinpayCallback) Handle(c *gin.Context) (err error) {
	if !s.IsValid() {
		err = errors.New("invalid response")
		return
	}
	// check api response
	// lock cash order
	// update cash order
	// {
	// update user_sum
	// create transaction history
	// }
	_, err = cashin.CloseCashInOrder(c, s.MerchantOrderNo, s.Amount, 0, 0, serializer.JSON(s), "", "", model.DB,0)
	if err != nil {
		return
	}
	return
}
