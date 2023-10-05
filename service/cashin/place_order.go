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
	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TopUpOrderService struct {
	Amount   string `form:"amount" json:"amount"`
	MethodID int64  `form:"method_id" json:"method_id"`
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
	r, err = s.verifyCashInAmount(c, amount)
	if err != nil {
		return
	}

	if amount < 0 {
		err = errors.New("illegal amount")
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
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
		r = serializer.EnsureErr(c, err, r)
		return
	}
	required, err := strconv.ParseBool(value)
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}
	if required {
		var kyc model.Kyc
		kyc, err = model.GetKycWithLock(model.DB, user.ID)
		if err != nil {
			r = serializer.EnsureErr(c, err, r)
			return
		}
		if kyc.Status != consts.KycStatusCompleted {
			err = errors.New("kyc not completed")
			r = serializer.EnsureErr(c, err, r)
			return
		}
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
		"")
	if err = model.DB.Debug().WithContext(c).Create(&cashOrder).Error; err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}
	var transactionID string
	switch s.MethodID {
	default:
		err = errors.New("unsupported method")
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	case 1:
		var data finpay.PaymentOrderRespData
		data, err = finpay.FinpayClient{}.PlaceDefaultGcashOrderV1(c, cashOrder.AppliedCashInAmount, 1, cashOrder.ID)
		if err != nil {
			r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
			return
		}
		transactionID = data.PaymentOrderNo
		r.Data = serializer.BuildPaymentOrder(data)
	}
	cashOrder.TransactionId = transactionID
	cashOrder.Status = models.CashOrderStatusPending
	_ = model.DB.Debug().WithContext(c).Save(&cashOrder)
	return
}

func (s TopUpOrderService) verifyCashInAmount(c *gin.Context, amount int64) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	if amount < 0 {
		err = errors.New("illegal amount")
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("negative_amount"), err)
		return
	}

	var firstTime bool = false
	err = model.DB.Where("user_id", user.ID).Where("order_type > 0").Where("status", models.CashOrderStatusSuccess).First(&model.CashOrder{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			firstTime = true
			err = nil
		} else {
			return
		}
	}

	minAmount := consts.TopupMinimum
	if firstTime {
		minAmount = consts.FirstTopupMinimum
	}
	if amount < minAmount {
		if firstTime {
			err = errors.New("illegal amount")
			r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("first_topup_amount"), err)
			return
		} else {
			err = errors.New("illegal amount")
			r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("topup_amount"), err)
			return
		}
	}

	return
}
