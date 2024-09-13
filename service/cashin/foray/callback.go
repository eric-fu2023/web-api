package cashin_foray

import (
	"errors"
	"web-api/model"
	"web-api/service/cashin"
	"web-api/util"

	"blgit.rfdev.tech/taya/payment-service/foray"
	"github.com/gin-gonic/gin"
)

type ForayPaymentCallback struct {
	foray.PaymentOrderCallbackReq
}

func (s *ForayPaymentCallback) Handle(c *gin.Context) (err error) {
	if !s.IsValid() {
		err = errors.New("invalid response")
		return
	}
	defer model.CashOrder{}.MarkCallbackAt(c, s.OrderTranoIn, model.DB)

	if !s.IsSuccess() {
		err = cashin.MarkOrderFailed(c, s.OrderTranoIn, util.JSON(s), s.OrderNumber)
		return
	}
	// check api response
	// lock cash order
	// update cash order
	// {
	// update user_sum
	// create transaction history
	// }
	_, err = cashin.CloseCashInOrder(c, s.OrderTranoIn, s.OrderAmount, 0, 0, util.JSON(s), model.DB, 10000)
	if err != nil {
		return
	}
	return
}
