package cashin_finpay

import (
	"encoding/json"
	"errors"
	"web-api/model"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


type FinpayCallback struct {
	finpay.PaymentOrderCallBackReq
}

func (s *FinpayCallback) Handle(c *gin.Context) (err error) {
	if !s.IsValid() {
		err = errors.New("invalid response")
		return
	}
	var order model.CashOrder
	// check api response
	// lock cash order
	// update cash order
	// {
	// update user_sum
	// create transaction history
	// }
	bytes,_ := json.Marshal(s)

	err = model.DB.Debug().WithContext(c).Transaction(func(tx *gorm.DB) (e error) {
		order, e = order.GetPendingWithLockWithDB(s.MerchantOrderNo, tx)
		if e != nil {
			return
		}
		order.Status = 2
		order.Notes = string(bytes)
		e = tx.Save(&order).Error
		if e != nil {
			return
		}
		
		return
	})
	return
}
