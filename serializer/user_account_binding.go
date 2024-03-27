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
	info := a.GetBankInfo()
	bytes, _ := json.Marshal(info)
	return UserAccountBinding{
		ID:            a.ID,
		UserID:        a.UserID,
		CashMethodID:  a.CashMethodID,
		AccountName:   string(a.AccountName),
		AccountNumber: string(a.AccountNumber),
		IsActive:      a.IsActive,
		BankInfo:      json.RawMessage(bytes),
		Method:        buildCashMethod(*a.CashMethod),
	}
}
