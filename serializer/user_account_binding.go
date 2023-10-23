package serializer

import "web-api/model"

type UserAccountBinding struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	CashMethodID  int64      `json:"cash_method_id"`
	AccountName   string     `json:"account_name"`
	AccountNumber string     `json:"account_number"`
	IsActive      bool       `json:"is_active"`
	Method        CashMethod `json:"method"`
}

func BuildUserAccountBinding(a model.UserAccountBinding) UserAccountBinding {
	return UserAccountBinding{
		ID:            a.ID,
		UserID:        a.UserID,
		CashMethodID:  a.CashMethodID,
		AccountName:   a.AccountName,
		AccountNumber: a.AccountNumber,
		IsActive:      a.IsActive,
		Method:        BuildCashMethod(*a.CashMethod),
	}
}
