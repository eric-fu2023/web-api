package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"sync"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"
)

type WalletService struct {
}

func (service *WalletService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	user := c.MustGet("user").(model.User)
	var gvu []ploutos.GameVendorUser
	err = model.DB.Model(ploutos.GameVendorUser{}).Scopes(model.GameVendorUserDefaultJoinAndPreload).
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
	err = model.DB.Model(ploutos.GameVendorUser{}).Scopes(model.GameVendorUserDefaultJoinAndPreload).
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
	Force bool `form:"force" json:"force"`
}

func (service *RecallFundService) Recall(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	lang := c.MustGet("_locale").(string)
	user := c.MustGet("user").(model.User)
	var userSum ploutos.UserSum
	err = model.DB.Where(`user_id`, user.ID).First(&userSum).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	if userSum.IsRecallNeeded || service.Force {
		var gvu []ploutos.GameVendorUser
		err = model.DB.Model(ploutos.GameVendorUser{}).Scopes(model.GameVendorUserDefaultJoinAndPreload).Where(`user_id`, user.ID).Find(&gvu).Error
		if err != nil {
			r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
			return
		}
		tx := model.DB.Begin()
		var wg sync.WaitGroup
		for _, g := range gvu {
			if g.GameVendor.GameIntegrationId == 0 {
				continue
			}
			wg.Add(1)
			go func(tx *gorm.DB, g ploutos.GameVendorUser) {
				defer wg.Done()
				err = common.GameIntegration[g.GameVendor.GameIntegrationId].TransferFrom(tx, user, g.ExternalCurrency, lang, g.GameVendor.GameCode, c.ClientIP())
				if err != nil {
					util.Log().Error("GAME INTEGRATION RECALL ERROR game_integration_id: %d, game_code: %s, user_id: %d, error: %s", g.GameVendor.GameIntegrationId, g.GameVendor.GameCode, user.ID, err.Error())
					return
				}
				err = tx.Model(ploutos.GameVendorUser{}).Where(`id`, g.ID).Updates(map[string]interface{}{"balance": 0, "is_last_played": false}).Error
				if err != nil {
					util.Log().Error("GAME INTEGRATION RECALL DB UPDATE ERROR game_integration_id: %d, game_code: %s, user_id: %d, error: %s", g.GameVendor.GameIntegrationId, g.GameVendor.GameCode, user.ID, err.Error())
					return
				}
			}(tx, g)
		}
		wg.Wait()
		tx.Commit()
		err = model.DB.Where(`user_id`, user.ID).First(&userSum).Error
		if err != nil {
			r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
			return
		}
		_ = model.DB.Model(ploutos.UserSum{}).Where(`user_id`, user.ID).Update(`is_recall_needed`, false).Error
	}
	r = serializer.Response{
		Data: serializer.BuildUserSum(userSum),
	}
	return
}
