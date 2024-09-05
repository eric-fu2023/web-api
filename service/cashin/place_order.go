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

	// convert amount from buck to cent
	amountDecimal, err := decimal.NewFromString(s.Amount)
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}
	// TODO: allow decimal places
	amount := amountDecimal.IntPart() * 100

	// retrieve channel
	method, err := model.CashMethod{}.GetByIDWithChannel(c, s.MethodID)
	if err != nil {
		return
	}
	cashMethodChannels := model.FilterByAmount(c, amount, model.FilterChannelByVip(c, user, method.CashMethodChannel))
	if len(cashMethodChannels) == 0 {
		err = errors.New("illegal amount")
		r = serializer.ParamErr(c, s, i18n.T("invalid_amount"), err)
		return
	}
	channel := model.GetNextChannel(cashMethodChannels)
	stats := channel.Stats

	// verify amount
	r, err = s.verifyCashInAmount(c, amount, method)
	if err != nil {
		return
	}

	if amount < 0 {
		err = errors.New("illegal amount")
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	// verify payment method
	err = verifyCashInMethod(method)
	if err != nil {
		return
	}

	// get exchange rate
	var exchangeClient exchange.ExchangeClient
	er, err := exchangeClient.GetExchangeRate(c, method.Currency, true)
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}

	// create CashInOrder
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

	// if currency is not USDT, round cashinAmount to nearest multiple 100, remove decimal
	// because most payment channels don't accept decimal places, except USDT wallet
	cashinAmount := int64(float64(cashOrder.AppliedCashInAmount) * er.AdjustedExchangeRate)
	if method.Currency != exchange.USDT && er.AdjustedExchangeRate != 1 && er.ExchangeRate != 1 {
		cashinAmount = (cashinAmount / 100) * 100
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

		switch config.Type {
		default:
			data, err = finpay.FinpayClient{}.PlaceDefaultOrderV1(c, cashinAmount, 1, user.ID, cashOrder.ID, config.Type, method.Currency, user.Username, "", config.TypeExtra)
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
			data, err = finpay.FinpayClient{}.PlaceDefaultOrderV1(c, cashinAmount, 1, user.ID, cashOrder.ID, config.Type, method.Currency, user.Username, "", json.RawMessage(raw))
			if err != nil {
				_ = MarkOrderFailed(c, cashOrder.ID, util.JSON(data), data.PaymentOrderNo)
				r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
				return
			}
		case "TRC20":
			data, err = finpay.FinpayClient{}.PlaceDefaultCoinPalOrderV1(c, cashinAmount, 1, user.ID, cashOrder.ID, user.Username)
			if err != nil {
				_ = MarkOrderFailed(c, cashOrder.ID, util.JSON(data), data.PaymentOrderNo)
				r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
				return
			}
		}
		transactionID = data.PaymentOrderNo
		r.Data = serializer.BuildPaymentOrder(data, method.Currency, decimal.NewFromInt(cashinAmount).Div(decimal.NewFromInt(100)))
		cashOrder.TransactionId = &transactionID
		cashOrder.Status = models.CashOrderStatusPending
	}
	_ = model.DB.Debug().WithContext(c).Save(&cashOrder)

	return
}

func (s TopUpOrderService) verifyCashInAmount(c *gin.Context, amount int64, method model.CashMethod) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	if amount < 0 || amount > method.MaxAmount || amount < method.MinAmount {
		err = errors.New("illegal amount")
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("negative_amount"), err)
		return
	}
	return
}

func verifyCashInMethod(m model.CashMethod) (err error) {
	if m.MethodType < 0 {
		return errors.New("cash method not permitted")
	}
	return nil
}
