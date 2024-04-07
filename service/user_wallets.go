package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"
)

type WalletService struct {
}

func (service *WalletService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	user := c.MustGet("user").(model.User)
	var gvu []ploutos.GameVendorUser
	err = model.DB.Model(ploutos.GameVendorUser{}).InnerJoins(`GameVendor`).Preload(`GameVendor.GameVendorBrand`).
		Where(`user_id`, user.ID).Where(`"GameVendor".game_integration_id != ?`, 0).Find(&gvu).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var data []serializer.Wallet
	for _, g := range gvu {
		data = append(data, serializer.BuildWallet(g))
	}
	r = serializer.Response{
		Data: data,
	}
	return
}

type SyncWalletService struct {
	GameCode string `form:"game_code" json:"game_code" binding:"required"`
}

func (service *SyncWalletService) Update(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	lang := c.MustGet("_locale").(string)
	user := c.MustGet("user").(model.User)
	var gvu ploutos.GameVendorUser
	err = model.DB.Model(ploutos.GameVendorUser{}).InnerJoins(`GameVendor`).Preload(`GameVendor.GameVendorBrand`).
		Where(`user_id`, user.ID).Where(`"GameVendor".game_code`, service.GameCode).First(&gvu).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	balance, err := common.GameIntegration[gvu.GameVendor.GameIntegrationId].GetGameBalance(user, gvu.ExternalCurrency, lang, gvu.GameVendor.GameCode, c.ClientIP())
	if gvu.Balance != balance {
		err = model.DB.Model(ploutos.GameVendorUser{}).Where(`id`, gvu.ID).Update(`balance`, balance).Error
		if err != nil {
			r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
			return
		}
		gvu.Balance = balance
	}
	r = serializer.Response{
		Data: serializer.BuildWallet(gvu),
	}
	return
}

type RecallFundService struct {
}

func (service *RecallFundService) Recall(c *gin.Context) (r serializer.Response, err error) {
	return
}
