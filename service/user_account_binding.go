package service

import (
	"encoding/json"
	"errors"
	"strings"
	"web-api/conf"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"

	models "blgit.rfdev.tech/taya/ploutos-object"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type ListWithdrawAccountsService struct {
}

func (s ListWithdrawAccountsService) List(c *gin.Context) (serializer.Response, error) {
	user := c.MustGet("user").(model.User)

	list, err := model.UserAccountBinding{}.GetAccountByUser(user.ID)
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err), nil
	}
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err), nil
	}

	return serializer.Response{
		Data: util.MapSlice(list, func(a model.UserAccountBinding) serializer.UserAccountBinding {
			return serializer.BuildUserAccountBinding(a, serializer.Modifier(serializer.BuildCashMethod, func(cm serializer.CashMethod) serializer.CashMethod {
				firstTopup, err := model.FirstTopup(c, user.ID)
				if err != nil || len(firstTopup.ID) == 0 {
					cm.MinAmount = conf.GetCfg().WithdrawMinNoDeposit / 100
				}
				return cm
			}))
		}),
	}, nil
}

type AddWithdrawAccountService struct {
	MethodID       int64  `form:"method_id" json:"method_id" binding:"required"`
	AccountNo      string `form:"account_no" json:"account_no" binding:"required"`
	AccountName    string `form:"account_name" json:"account_name" binding:"required"`
	BankCode       string `form:"bank_code" json:"bank_code"`
	BankBranchName string `form:"bank_branch_name" json:"bank_branch_name"`
	BankName       string `form:"bank_name" json:"bank_name"`
}

func (s AddWithdrawAccountService) Do(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	user := c.MustGet("user").(model.User)
	// r, err = VerifyKycWithName(c, user.ID, s.AccountName)
	// if err != nil {
	// 	return
	// }
	method, err := model.CashMethod{}.GetByID(c, s.MethodID, int(user.BrandId))
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err), err
	}
	s.AccountNo = strings.TrimLeft(s.AccountNo, "+")
	switch method.AccountType {
	case "crypto_wallet_trc20":
		if !strings.HasPrefix(s.AccountNo, "T") || len(s.AccountNo) != 34 {
			err = errors.New("wrong_format")
			return serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("invalid_account_format"), err), err
		}
	}

	accountBinding := model.UserAccountBinding{
		UserAccountBinding: ploutos.UserAccountBinding{
			UserID:        user.ID,
			CashMethodID:  s.MethodID,
			AccountName:   s.AccountName,
			AccountNumber: models.EncryptedStr(s.AccountNo),
			IsActive:      true,
		},
	}
	accountBinding.SetBankInfo(ploutos.BankInfo{
		BankCode:       s.BankCode,
		BankBranchName: s.BankBranchName,
		BankName:       s.BankName,
	})
	err = accountBinding.AddToDb()
	if err != nil {
		if errors.Is(err, model.ErrAccountLimitExceeded) {
			r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("withdraw_account_limit_exceeded"), err)
			return
		}
		r = serializer.Err(c, s, serializer.CodeGeneralError, "", err)
		return
	}
	return
}

type DeleteWithdrawAccountService struct {
	AccountBindingID json.Number `form:"account_binding_id" json:"account_binding_id" binding:"required"`
}

func (s DeleteWithdrawAccountService) Do(c *gin.Context) (r serializer.Response, err error) {
	user := c.MustGet("user").(model.User)
	i18n := c.MustGet("i18n").(i18n.I18n)

	accID, _ := s.AccountBindingID.Int64()

	accountBinding := model.UserAccountBinding{
		UserAccountBinding: ploutos.UserAccountBinding{
			BASE:     ploutos.BASE{ID: accID},
			UserID:   user.ID,
			IsActive: true,
		},
	}
	err = accountBinding.Remove()
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err), err
	}
	return serializer.Response{Msg: i18n.T("success")}, nil
}
