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
	defer model.CashOrder{}.MarkCallbackAt(c, s.MerchantOrderNo, model.DB)

	if !s.IsSuccess() {
		err = cashin.MarkOrderFailed(c, s.MerchantOrderNo, util.JSON(s), s.PaymentOrderNo)
		return
	}
	// check api response
	// lock cash order
	// update cash order
	// {
	// update user_sum
	// create transaction history
	// }
	_, err = cashin.CloseCashInOrder(c, s.MerchantOrderNo, s.Amount, 0, 0, util.JSON(s), model.DB, 10000)
	if err != nil {
		return
	}
	return
}
