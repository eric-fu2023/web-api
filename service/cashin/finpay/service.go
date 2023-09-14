package cashin_finpay

import (
	"errors"
	"fmt"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/cashin"
	"web-api/util/i18n"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	"github.com/gin-gonic/gin"
)

type TopUpOrderService struct {
	Amount   int64  `form:"amount" json:"amount"`
	Quantity int64  `form:"quantity" json:"quantity"`
	Sku      string `form:"sku" json:"sku"`
	MethodID int64  `form:"method_id" json:"method_id"`
}

func (s *TopUpOrderService) CreateOrder(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var cl finpay.FinpayClient
	fmt.Print(user)
	fmt.Print(cl)
	// check kyc
	// create cash order
	// create payment order
	// err handling and return
	var kyc model.Kyc
	kyc ,err = model.GetKycWithLock(model.DB,user.ID)
	if err != nil {
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	if kyc.Status != consts.KycStatusCompleted {
		err = errors.New("kyc not completed")
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	var userSum model.UserSum
	userSum,err = model.UserSum{}.GetByIDWithLockWithDB(user.ID,model.DB)
	if err != nil {
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	cashOrder := model.NewCashInOrder(user.ID, 
		s.MethodID, 
		s.Amount*100, 
		userSum.Balance,
		cashin.GetWagerFromAmount(s.Amount,cashin.DefaultWager), 
		"")
	var data finpay.PaymentOrderRespData
	if err = model.DB.Debug().WithContext(c).Create(&cashOrder).Error; err != nil {
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	if data, err = cl.PlaceDefaultGcashOrderV1(c, cashOrder.AppliedCashInAmount, 1, cashOrder.ID); err != nil {
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	r.Data = serializer.BuildPaymentOrder(data)
	return
}
