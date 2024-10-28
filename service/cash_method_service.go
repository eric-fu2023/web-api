package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"web-api/conf"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/promotion/cash_method_promotion"
	"web-api/util"
	"web-api/util/i18n"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type CasheMethodListService struct {
	WithdrawOnly bool `form:"withdraw_only" json:"withdraw_only"`
	TopupOnly    bool `form:"topup_only" json:"topup_only"`
}

func (s CasheMethodListService) List(c *gin.Context) (serializer.Response, error) {
	brand := c.MustGet(`_brand`).(int)

	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user, _ := u.(model.User)
	deviceInfo, _ := util.GetDeviceInfo(c)

	vip, err := model.GetVipWithDefault(c, user.ID)
	if err != nil {
		r := serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return r, err
	}

	ctx := rfcontext.AppendCallDesc(rfcontext.Spawn(context.Background()), "CasheMethodListService.List")

	{
		ctx = rfcontext.AppendParams(ctx, "CasheMethodListService.List", map[string]interface{}{
			"brand":     brand,
			"withdraw?": s.WithdrawOnly,
			"deposit?":  s.TopupOnly,
			"vip_id":    vip.VipID,
		})
	}
	var cashMethods []model.CashMethod
	cashMethods, err = model.CashMethod{}.List(c, s.WithdrawOnly, s.TopupOnly, deviceInfo.Platform, brand, int(vip.VipRule.ID), user)
	if err != nil {
		r := serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return r, err
	}

	weeklyAmountRecords, dailyAmountRecords, err := cash_method_promotion.GetAccumulatedClaimedCashMethodPromotionPast7And1Days(c, 0, user.ID)
	maxPromotionAmountByCashMethodId := map[ /*cash_method.id*/ int64] /*max amount*/ int64{}
	util.MapSlice(cashMethods, func(a model.CashMethod) (err error) {
		if a.CashMethodPromotion == nil {
			return
		}
		weeklyAmount := util.FindOrDefault(weeklyAmountRecords, func(b ploutos.CashMethodPromotionRecord) bool {
			return b.CashMethodId == a.ID
		}).Amount
		dailyAmount := util.FindOrDefault(dailyAmountRecords, func(b ploutos.CashMethodPromotionRecord) bool {
			return b.CashMethodId == a.ID
		}).Amount

		maxAmount, err := cash_method_promotion.FinalPayout(c, weeklyAmount, dailyAmount, *a.CashMethodPromotion, 0, true)
		if err != nil {
			util.GetLoggerEntry(c).Error("HandleCashMethodPromotion GetMaxAmountPayment", err)
		}
		maxPromotionAmountByCashMethodId[a.ID] = maxAmount
		return
	})

	var r serializer.Response
	if s.TopupOnly {
		r.Data = util.MapSlice(cashMethods, func(a model.CashMethod) serializer.CashMethod {
			return serializer.BuildCashMethod(a, maxPromotionAmountByCashMethodId)
		})
	} else {
		r.Data = util.MapSlice(cashMethods, serializer.Modifier(
			func(a model.CashMethod) serializer.CashMethod {
				return serializer.BuildCashMethod(a, maxPromotionAmountByCashMethodId)
			},
			func(cm serializer.CashMethod) serializer.CashMethod {
				firstTopup, err := model.FirstTopup(c, user.ID)
				if err != nil || len(firstTopup.ID) == 0 {
					cm.MinAmount = max(conf.GetCfg().WithdrawMinNoDeposit/100, cm.MinAmount)
				}
				return cm
			}))
	}

	responseB, _ := json.Marshal(r)
	ctx = rfcontext.AppendDescription(ctx, fmt.Sprintf("response %s", string(responseB)))

	go log.Println(rfcontext.Fmt(ctx))
	return r, nil
}
