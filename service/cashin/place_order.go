package cashin

import (
	"errors"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type TopUpOrderService struct {
	Amount   string `form:"amount" json:"amount"`
	MethodID int64  `form:"method_id" json:"method_id"`
}

func (s TopUpOrderService) CreateOrder(c *gin.Context) (r serializer.Response, err error) {
	brand := c.MustGet(`_brand`).(int)
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)
	method, err := model.CashMethod{}.GetByID(c, s.MethodID, brand)
	if err != nil {
		return
	}

	amountDecimal, err := decimal.NewFromString(s.Amount)
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}
	amount := amountDecimal.IntPart() * 100
	r, err = s.verifyCashInAmount(c, amount, method)
	if err != nil {
		return
	}

	if amount < 0 {
		err = errors.New("illegal amount")
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	err = processCashInMethod(method)
	if err != nil {
		return
	}

	// switch s.MethodID {
	// case 1, 8:
	// default:
	// 	err = errors.New("unsupported method")
	// 	r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
	// 	return
	// }

	// check kyc
	// create cash order
	// create payment order
	// err handling and return
	// _, err = service.VerifyKyc(c, user.ID)
	// if err != nil {
	// 	r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("kyc_get_failed"), err)
	// 	return
	// }
	var userSum model.UserSum
	userSum, err = model.UserSum{}.GetByUserIDWithLockWithDB(user.ID, model.DB)
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}
	cashOrder := model.NewCashInOrder(user.ID,
		s.MethodID,
		amount,
		userSum.Balance,
		GetWagerFromAmount(amount, DefaultWager),
		c.ClientIP())
	if err = model.DB.Debug().WithContext(c).Create(&cashOrder).Error; err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}
	var transactionID string
	switch method.Gateway {
	default:
		err = errors.New("unsupported method")
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	case "finpay":
		config := method.GetFinpayConfig()
		var data finpay.PaymentOrderRespData
		switch config.Type {
		default:
			data, err = finpay.FinpayClient{}.PlaceDefaultOrderV1(c, cashOrder.AppliedCashInAmount, 1, cashOrder.ID, config.Type, method.Currency, user.Username)
			if err != nil {
				_ = MarkOrderFailed(c, cashOrder.ID, util.JSON(data))
				r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
				return
			}
		case "TRC20":
			data, err = finpay.FinpayClient{}.PlaceDefaultCoinPalOrderV1(c, cashOrder.AppliedCashInAmount, 1, cashOrder.ID, user.Username)
			if err != nil {
				_ = MarkOrderFailed(c, cashOrder.ID, util.JSON(data))
				r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
				return
			}
		}
		transactionID = data.PaymentOrderNo
		r.Data = serializer.BuildPaymentOrder(data)
		cashOrder.TransactionId = &transactionID
		cashOrder.Status = models.CashOrderStatusPending
	}
	_ = model.DB.Debug().WithContext(c).Save(&cashOrder)
	return
}

func (s TopUpOrderService) verifyCashInAmount(c *gin.Context, amount int64, method model.CashMethod) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	// u, _ := c.Get("user")
	// user := u.(model.User)

	if amount < 0 {
		err = errors.New("illegal amount")
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("negative_amount"), err)
		return
	}
	return

	// var firstTime bool = false
	// err = model.DB.WithContext(c).Where("user_id", user.ID).Where("order_type > 0").Where("status", models.CashOrderStatusSuccess).First(&model.CashOrder{}).Error
	// if err != nil {
	// 	if errors.Is(err, gorm.ErrRecordNotFound) {
	// 		firstTime = true
	// 		err = nil
	// 	} else {
	// 		return
	// 	}
	// }

	// minAmount := method.MinAmount
	// if amount < minAmount {
	// 	if firstTime {
	// 		err = errors.New("illegal amount")
	// 		r = serializer.Err(c, s, serializer.CodeGeneralError, fmt.Sprintf(i18n.T("first_topup_amount"), minAmount/100), err)
	// 		return
	// 	} else {
	// 		err = errors.New("illegal amount")
	// 		r = serializer.Err(c, s, serializer.CodeGeneralError, fmt.Sprintf(i18n.T("topup_amount"), minAmount/100), err)
	// 		return
	// 	}
	// }

	return
}

func processCashInMethod(m model.CashMethod) (err error) {
	if m.MethodType < 0 {
		return errors.New("cash method not permitted")
	}
	return nil
}
