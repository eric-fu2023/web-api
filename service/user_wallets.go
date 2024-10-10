package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"blgit.rfdev.tech/taya/common-function/rfcontext"

	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
	"github.com/go-redsync/redsync/v4"
)

const (
	userWalletRecallKey = "user_wallet_recall_lock:%d"
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
	locale := c.MustGet("_locale").(string)
	user := c.MustGet("user").(model.User)
	var gvu ploutos.GameVendorUser
	err = model.DB.Model(ploutos.GameVendorUser{}).Scopes(model.GameVendorUserDefaultJoinAndPreload).
		Where(`user_id`, user.ID).Where(`"GameVendor".game_code`, service.GameCode).First(&gvu).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	extra := model.Extra{Locale: locale, Ip: c.ClientIP()}
	balance, err := common.GameIntegration[gvu.GameVendor.GameIntegrationId].GetGameBalance(context.TODO(), user, gvu.ExternalCurrency, gvu.GameVendor.GameCode, extra)
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
	locale := c.MustGet("_locale").(string)
	user := c.MustGet("user").(model.User)
	userSum, err := recall(user, service.Force, locale, c.ClientIP())
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	r = serializer.Response{
		Data: serializer.BuildUserSum(userSum),
	}
	return
}

type InternalRecallFundService struct {
	UserId int64  `form:"user_id" json:"user_id"`
	Locale string `form:"locale" json:"locale"`
}

func (service *InternalRecallFundService) Recall(c *gin.Context) (r serializer.Response, err error) {
	var user model.User
	err = model.DB.Where(`id`, service.UserId).First(&user).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, "user not found", err)
		return
	}
	userSum, err := recall(user, true, service.Locale, c.ClientIP())
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, "something went wrong", err)
		return
	}
	r = serializer.Response{
		Data: serializer.BuildUserSum(userSum),
	}
	return
}

func recall(user model.User, force bool, locale, ip string) (userSum ploutos.UserSum, err error) {
	ctx := rfcontext.AppendCallDesc(rfcontext.Spawn(context.Background()), "recall")

	mutex := cache.RedisLockClient.NewMutex(fmt.Sprintf(userWalletRecallKey, user.ID), redsync.WithExpiry(50*time.Second))
	mutex.Lock()
	defer mutex.Unlock()
	err = model.DB.Where(`user_id`, user.ID).First(&userSum).Error
	if err != nil {
		return
	}
	if userSum.IsRecallNeeded || force {
		var gvu []ploutos.GameVendorUser
		err = model.DB.Model(ploutos.GameVendorUser{}).Scopes(model.GameVendorUserDefaultJoinAndPreload).Where(`user_id`, user.ID).Find(&gvu).Error
		if err != nil {
			return
		}
		var wg sync.WaitGroup
		for _, g := range gvu {
			if g.GameVendor.GameIntegrationId == 0 {
				continue
			}
			wg.Add(1)

			ctx = rfcontext.AppendStats(ctx, "game_vendor_found", 1)
			go func(g ploutos.GameVendorUser) {
				rCtx := ctx

				defer wg.Done()
				tx := model.DB.Begin()
				defer tx.Rollback()

				extra := model.Extra{Locale: locale, Ip: ip}
				err = common.GameIntegration[g.GameVendor.GameIntegrationId].TransferFrom(rCtx, tx, user, g.ExternalCurrency, g.GameVendor.GameCode, g.GameVendorId, extra)
				if err != nil {
					util.Log().Error(rfcontext.Fmt(rfcontext.AppendError(rCtx, err, fmt.Sprintf("`GAME INTEGRATION RECALL ERROR game_integration_id: %d, game_code: %s, user_id: %d, error: %s", g.GameVendor.GameIntegrationId, g.GameVendor.GameCode, user.ID, err.Error()))))
					return
				}
				var maxRetries = 3
				for i := 0; i < maxRetries; i++ {
					err = tx.Model(ploutos.GameVendorUser{}).Where(`id`, g.ID).Updates(map[string]interface{}{"balance": 0, "is_last_played": false}).Error
					if err != nil {
						util.Log().Error(rfcontext.Fmt(rfcontext.AppendError(rCtx, err, fmt.Sprintf("GAME INTEGRATION RECALL DB UPDATE ERROR game_integration_id: %d, game_code: %s, user_id: %d, error: %s", g.GameVendor.GameIntegrationId, g.GameVendor.GameCode, user.ID, err.Error()))))
						if i == maxRetries-1 {
							ctx = rfcontext.AppendStats(ctx, "game_vendor_withdraw_process_db_fail_retry", 1)
							return
						}
						time.Sleep(200 * time.Millisecond)
						continue
					}
					break
				}

				tx.Commit()
			}(g)
		}
		wg.Wait()
		err = model.DB.Where(`user_id`, user.ID).First(&userSum).Error
		if err != nil {
			return
		}
		_ = model.DB.Model(ploutos.UserSum{}).Where(`user_id`, user.ID).Update(`is_recall_needed`, false).Error

		ctx = rfcontext.AppendCallDesc(ctx, "END")
	}
	return
}
