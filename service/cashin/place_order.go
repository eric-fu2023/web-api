package cashin

import (
	"errors"
	"strconv"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service"
	"web-api/util/i18n"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	"github.com/gin-gonic/gin"
)

type TopUpOrderService struct {
	Amount   int64 `form:"amount" json:"amount"`
	MethodID int64 `form:"method_id" json:"method_id"`
}

func (s TopUpOrderService) CreateOrder(c *gin.Context) (r serializer.Response, err error) {
	amount := s.Amount * 100

	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)
	switch s.MethodID {
	case 1:
	default:
		err = errors.New("unsupported method")
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	// check kyc
	// create cash order
	// create payment order
	// err handling and return
	value, err := service.GetCachedConfig(c, consts.ConfigKeyTopupKycCheck)
	if err != nil {
		return
	}
	required, err := strconv.ParseBool(value)
	if err != nil {
		return
	}
	if required {
		var kyc model.Kyc
		kyc, err = model.GetKycWithLock(model.DB.Debug(), user.ID)
		if err != nil {
			return
		}
		if kyc.Status != consts.KycStatusCompleted {
			err = errors.New("kyc not completed")
			return
		}
	}

	var userSum model.UserSum
	userSum, err = model.UserSum{}.GetByUserIDWithLockWithDB(user.ID, model.DB)
	if err != nil {
		return
	}
	cashOrder := model.NewCashInOrder(user.ID,
		s.MethodID,
		amount,
		userSum.Balance,
		GetWagerFromAmount(amount, DefaultWager),
		"")
	if err = model.DB.Debug().WithContext(c).Create(&cashOrder).Error; err != nil {
		return
	}
	var transactionID string
	switch s.MethodID {
	case 1:
		var data finpay.PaymentOrderRespData
		data, err = finpay.FinpayClient{}.PlaceDefaultGcashOrderV1(c, cashOrder.AppliedCashInAmount, 1, cashOrder.ID)
		if err != nil {
			r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
			return
		}
		transactionID = data.PaymentOrderNo
		r.Data = serializer.BuildPaymentOrder(data)
	default:
		err = errors.New("unsupported method")
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	cashOrder.TransactionId = transactionID
	_ = model.DB.Debug().WithContext(c).Save(&cashOrder)
	return
}
