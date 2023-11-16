package service

import (
	"encoding/json"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"

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
		Data: util.MapSlice(list, serializer.BuildUserAccountBinding),
	}, nil
}

type AddWithdrawAccountService struct {
	MethodID    int64  `form:"method_id" json:"method_id" binding:"required"`
	AccountNo   string `form:"account_no" json:"account_no" binding:"required"`
	AccountName string `form:"account_name" json:"account_name" binding:"required"`
}

func (s AddWithdrawAccountService) Do(c *gin.Context) (serializer.Response, error) {
	user := c.MustGet("user").(model.User)
	i18n := c.MustGet("i18n").(i18n.I18n)
	r, err := VerifyKycWithName(c, user.ID, s.AccountName)
	if err != nil {
		return r, err
	}

	accountBinding := model.UserAccountBinding{
		UserAccountBinding: ploutos.UserAccountBinding{
			UserID:        user.ID,
			CashMethodID:  s.MethodID,
			AccountName:   s.AccountName,
			AccountNumber: s.AccountNo,
			IsActive:      true,
		},
	}
	err = accountBinding.AddToDb()
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err), err
	}
	return serializer.Response{Msg: i18n.T("success")}, nil
}

type DeleteWithdrawAccountService struct {
	AccountBindingID json.Number `form:"account_binding_id" json:"account_binding_id" binding:"required"`
}

func (s DeleteWithdrawAccountService) Do(c *gin.Context) (serializer.Response, error) {
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
	err := accountBinding.Remove()
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err), err
	}
	return serializer.Response{Msg: i18n.T("success")}, nil
}
