package cashin

import (
	"encoding/json"
	"errors"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/exchange"
	"web-api/util"
	"web-api/util/i18n"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type TopUpOrderService struct {
	Amount          string `form:"amount" json:"amount"`
	MethodID        int64  `form:"method_id" json:"method_id"`
	BankAccountName string `form:"bank_account_name" json:"bank_account_name"`
}

func (s TopUpOrderService) CreateOrder(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	amountDecimal, err := decimal.NewFromString(s.Amount)
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}
	amount := amountDecimal.IntPart() * 100

	method, err := model.CashMethod{}.GetByIDWithChannel(c, s.MethodID)
	if err != nil {
		return
	}
	channel := model.GetNextChannel(model.FilterByAmount(c, amount, model.FilterChannelByVip(c, user, method.CashMethodChannel)))
	stats := channel.Stats

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
	var exchangeClient exchange.OkxClient
	er, err := exchangeClient.GetExchangeRate(c, method.Currency, true)
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}
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
		c.ClientIP(), method.Currency, er.ExchangeRate, er.AdjustedExchangeRate)
	if err = model.DB.Debug().WithContext(c).Create(&cashOrder).Error; err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}
	var transactionID string
	switch channel.Gateway {
	default:
		err = errors.New("unsupported method")
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	case "finpay":
		config := channel.GetFinpayConfig()
		var data finpay.PaymentOrderRespData
		defer func() {
			result := "success"
			if errors.Is(err, finpay.ErrorGateway) {
				result = "gateway_failed"
			}
			if data.IsFailed() {
				result = "failed"
			}
			_ = model.IncrementStats(stats, result)

		}()
		cashinAmount := int64(float64(cashOrder.AppliedCashInAmount) * er.AdjustedExchangeRate)
		switch config.Type {
		default:
			data, err = finpay.FinpayClient{}.PlaceDefaultOrderV1(c, cashinAmount, 1, cashOrder.ID, config.Type, method.Currency, user.Username, "", config.TypeExtra)
			if err != nil {
				_ = MarkOrderFailed(c, cashOrder.ID, util.JSON(data), data.PaymentOrderNo)
				r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
				return
			}
		case "BANK_CARD", "BANK_CARD_H5":
			extra := map[string]string{
				"bankAccountName": s.BankAccountName,
			}
			raw, _ := json.Marshal(extra)
			data, err = finpay.FinpayClient{}.PlaceDefaultOrderV1(c, cashinAmount, 1, cashOrder.ID, config.Type, method.Currency, user.Username, "", json.RawMessage(raw))
			if err != nil {
				_ = MarkOrderFailed(c, cashOrder.ID, util.JSON(data), data.PaymentOrderNo)
				r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
				return
			}
		case "TRC20":
			data, err = finpay.FinpayClient{}.PlaceDefaultCoinPalOrderV1(c, cashinAmount, 1, cashOrder.ID, user.Username)
			if err != nil {
				_ = MarkOrderFailed(c, cashOrder.ID, util.JSON(data), data.PaymentOrderNo)
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

	if amount < 0 || amount > method.MaxAmount || amount < method.MinAmount {
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

}

func processCashInMethod(m model.CashMethod) (err error) {
	if m.MethodType < 0 {
		return errors.New("cash method not permitted")
	}
	return nil
}
