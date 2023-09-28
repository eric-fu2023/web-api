package cashin_finpay

import (
	"errors"
	"web-api/model"
	"web-api/service/cashin"
	"web-api/util"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	"github.com/gin-gonic/gin"
)

type FinpayPaymentCallback struct {
	finpay.PaymentOrderCallBackReq
}

func (s *FinpayPaymentCallback) Handle(c *gin.Context) (err error) {
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
	_, err = cashin.CloseCashInOrder(c, s.MerchantOrderNo, s.Amount, 0, 0, util.JSON(s), model.DB)
	if err != nil {
		return
	}
	return
}
