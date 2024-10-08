package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"

	"web-api/conf"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/promotion/cash_method_promotion"
	"web-api/util"
	"web-api/util/i18n"

	"blgit.rfdev.tech/taya/common-function/crypto/md5"
	"blgit.rfdev.tech/taya/common-function/rfcontext"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type ListWithdrawAccountsService struct {
}

func (s ListWithdrawAccountsService) List(c *gin.Context) (serializer.Response, error) {
	user := c.MustGet("user").(model.User)

	vip, err := model.GetVipWithDefault(c, user.ID)
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err), nil
	}

	list, err := model.UserAccountBinding{}.GetAccountByUser(c, user.ID, vip.VipRule.ID)
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err), nil
	}

	weeklyAmountRecords, dailyAmountRecords, err := cash_method_promotion.GetAccumulatedClaimedCashMethodPromotionPast7And1Days(c, 0, user.ID)
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err), nil
	}
	maxPromotionAmountByCashMethodId := map[int64]int64{}
	util.MapSlice(list, func(a model.UserAccountBinding) (err error) {
		if a.CashMethod == nil {
			return
		}
		if a.CashMethod.CashMethodPromotion == nil {
			return
		}
		weeklyAmount := util.FindOrDefault(weeklyAmountRecords, func(b ploutos.CashMethodPromotionRecord) bool {
			return b.CashMethodId == a.CashMethod.ID
		}).Amount
		dailyAmount := util.FindOrDefault(dailyAmountRecords, func(b ploutos.CashMethodPromotionRecord) bool {
			return b.CashMethodId == a.CashMethod.ID
		}).Amount

		maxAmount, err := cash_method_promotion.FinalPayout(c, weeklyAmount, dailyAmount, *a.CashMethod.CashMethodPromotion, 0, true)
		if err != nil {
			util.GetLoggerEntry(c).Error("HandleCashMethodPromotion GetMaxAmountPayment", err)
		}
		maxPromotionAmountByCashMethodId[a.CashMethod.ID] = maxAmount
		return
	})

	return serializer.Response{
		Data: util.MapSlice(list, func(a model.UserAccountBinding) serializer.UserAccountBinding {
			return serializer.BuildUserAccountBinding(a, serializer.Modifier(func(b model.CashMethod) serializer.CashMethod {
				return serializer.BuildCashMethod(b, maxPromotionAmountByCashMethodId)
			}, func(cm serializer.CashMethod) serializer.CashMethod {
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
	FirstName      string `form:"first_name" json:"first_name"`
	LastName       string `form:"last_name" json:"last_name"`
}

func (s AddWithdrawAccountService) Do(c *gin.Context) (r serializer.Response, err error) {
	ctx := rfcontext.AppendCallDesc(rfcontext.Spawn(context.Background()), "AddWithdrawAccountService")
	i18n := c.MustGet("i18n").(i18n.I18n)
	user := c.MustGet("user").(model.User)
	ctx = rfcontext.AppendParams(ctx, "AddWithdrawAccountService", map[string]interface{}{
		"middleware_user": user,
		"params":          s,
	})
	// r, err = VerifyKycWithName(c, user.ID, s.AccountName)
	// if err != nil {
	// 	return
	// }
	method, err := model.CashMethod{}.GetByID(c, s.MethodID, int(user.BrandId))
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err), err
	}
	accountNo := strings.TrimLeft(s.AccountNo, "+")

	// validate account number format
	switch method.AccountType {
	case "crypto_wallet_trc20":
		if !strings.HasPrefix(s.AccountNo, "T") || len(s.AccountNo) != 34 {
			err = errors.New("wrong_format")
			return serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("invalid_account_format"), err), err
		}
	}

	// hash
	hashSalt := os.Getenv("MOBILE_EMAIL_HASH_SALT")
	service, err := md5.NewWithSalt(hashSalt)
	if err != nil {
		ctx = rfcontext.AppendError(ctx, err, "NewWithSalt")
		log.Printf(rfcontext.Fmt(ctx))
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("500"), err)
		return
	}

	accountNoHash := service.Hash([]byte(accountNo))

	accountBinding := model.UserAccountBinding{
		UserAccountBinding: ploutos.UserAccountBinding{
			UserID:            user.ID,
			CashMethodID:      s.MethodID,
			AccountName:       ploutos.EncryptedStr(s.AccountName),
			AccountNumber:     ploutos.EncryptedStr(accountNo),
			AccountNumberHash: accountNoHash,
			IsActive:          true,
		},
	}

	if method.AccountType == "paypal_email" {
		accountBinding.AccountName = "PayPal"
	}

	accountBinding.SetBankInfo(ploutos.BankInfo{
		BankCode:       s.BankCode,
		BankBranchName: s.BankBranchName,
		BankName:       s.BankName,
		FirstName:      s.FirstName,
		LastName:       s.LastName,
	})

	err = accountBinding.AddToDb()
	if err != nil {
		ctx = rfcontext.AppendError(ctx, err, "AddToDb")
		log.Printf(rfcontext.Fmt(ctx))

		if errors.Is(err, model.ErrAccountLimitExceeded) {
			r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("withdraw_account_limit_exceeded"), err)
			return
		}
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("withdrawal_account_already_used"), err)
		return
	}

	log.Printf(rfcontext.Fmt(ctx))
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
	err = accountBinding.HardRemove()
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err), err
	}
	return serializer.Response{Msg: i18n.T("success")}, nil
}
