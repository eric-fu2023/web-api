package serializer

import (
	"encoding/json"
	"web-api/model"
)

type UserAccountBinding struct {
	ID            int64           `json:"id"`
	UserID        int64           `json:"user_id"`
	CashMethodID  int64           `json:"cash_method_id"`
	AccountName   string          `json:"account_name"`
	AccountNumber string          `json:"account_number"`
	IsActive      bool            `json:"is_active"`
	BankInfo      json.RawMessage `json:"bank_info"`
	Method        CashMethod      `json:"method"`
}

func BuildUserAccountBinding(a model.UserAccountBinding, buildCashMethod func(a model.CashMethod) CashMethod) UserAccountBinding {
	return UserAccountBinding{
		ID:            a.ID,
		UserID:        a.UserID,
		CashMethodID:  a.CashMethodID,
		AccountName:   a.AccountName,
		AccountNumber: a.AccountNumber,
		IsActive:      a.IsActive,
		BankInfo:      a.BankInfo,
		Method:        buildCashMethod(*a.CashMethod),
	}
}
