package serializer

import (
	"web-api/conf/consts"
	"web-api/model"
	"web-api/util/i18n"

	"github.com/shopspring/decimal"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type TopupOrder struct {
	TopupOrderNo     string              `json:"topup_order_no"`
	TopupOrderStatus string              `json:"topup_order_status"`
	OrderNumber      string              `json:"order_number"`
	FromCurrency     string              `json:"from_currency"`
	FromAmount       decimal.Decimal     `json:"from_amount"`
	FromExchangeRate float64             `json:"from_exchange_rate"`
	ToCurrency       string              `json:"to_currency"`
	ToAmount         decimal.Decimal     `json:"to_amount"`
	ToExchangeRate   float64             `json:"to_exchange_rate"`
	TopupData        *string             `json:"topup_data"`
	TopupDataType    *string             `json:"topup_data_type"`
	RedirectUrl      string              `json:"redirect_url"`
	Html             string              `json:"html"`
	WalletAddress    string              `json:"wallet_address"`
	BankCardInfo     paymentBankCardInfo `json:"bank_card_info"`
}

type paymentBankCardInfo struct {
	Amount          float64 `json:"amount"`
	BankCode        string  `json:"bankCode"`
	BankName        string  `json:"bankName"`
	BankAccountNo   string  `json:"bankAccountNo"`
	BankBranchName  string  `json:"bankBranchName"`
	BankAccountName string  `json:"bankAccountName"`
}

type WithdrawOrder struct {
	WithdrawOrderNo     string `json:"withdraw_order_no"`
	WithdrawOrderStatus string `json:"withdraw_order_status"`
	OrderNumber         string `json:"order_number"`
}

func BuildWithdrawOrder(p ploutos.CashOrder) WithdrawOrder {
	txnID := ""
	if p.TransactionId != nil {
		txnID = *p.TransactionId
	}
	return WithdrawOrder{
		WithdrawOrderNo:     txnID,
		WithdrawOrderStatus: consts.CashOrderStatus[p.Status],
		OrderNumber:         p.ID,
	}
}

type GenericCashOrder struct {
	OrderNo         string  `json:"order_no"`
	OrderStatus     string  `json:"order_status"`
	CreatedAt       int64   `json:"created_at"`
	Amount          float64 `json:"amount"`
	EffectiveAmount float64 `json:"effective_amount"`
	OrderType       string  `json:"order_type"`
	// Currency    string    `json:"currency"`
	TypeDetail     string  `json:"type_detail"`
	Wager          float64 `json:"wager"`
	ErrReasonTitle string  `json:"err_reason_title"`
	ErrReasonDesc  string  `json:"err_reason_desc"`
}

func BuildGenericCashOrder(p model.CashOrder, i18n i18n.I18n) GenericCashOrder {
	amount := p.AppliedCashInAmount
	effectiveAmount := p.EffectiveCashInAmount
	orderType := consts.OrderTypeMap[p.OrderType]
	detail := i18n.T(consts.OrderOperationTypeDetailMap[p.OperationType])
	if p.OperationType == 0 {
		detail = i18n.T(consts.OrderTypeDetailMap[p.OrderType])
	}
	if p.OrderType < 0 {
		amount = p.AppliedCashOutAmount
		effectiveAmount = p.EffectiveCashOutAmount
	}

	if p.Name != "" {
		detail = p.Name
	}

	return GenericCashOrder{
		OrderNo:         p.ID,
		OrderStatus:     consts.CashOrderStatus[p.Status],
		CreatedAt:       p.CreatedAt.Unix(),
		Amount:          float64(amount) / 100,
		EffectiveAmount: float64(effectiveAmount) / 100,
		OrderType:       orderType,
		TypeDetail:      detail,
		Wager:           float64(p.WagerChange) / 100,
	}
}
