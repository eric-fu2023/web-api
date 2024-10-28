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
			"user":      user.ID,
			"vip_id":    vip.VipID,
		})
	}
	cashMethods, err := model.CashMethodWithPromotions(c, s.WithdrawOnly, s.TopupOnly, deviceInfo.Platform, brand, int(vip.VipRule.ID), user)
	if err != nil {
		r := serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return r, err
	}

	claimedPast7DaysL, claimedPast1DayL, err := cash_method_promotion.GetAccumulatedClaimedCashMethodPromotionPast7And1DaysM(c, user.ID)

	cashMethodsR := make([]serializer.CashMethod, 0, len(cashMethods))
	for _, cm := range cashMethods {
		cCtx := rfcontext.AppendParams(ctx, "moulding ", map[string]interface{}{
			"cash_method_id": cm.ID,
		})
		cashMethodId := cm.ID

		claimedPast7Days, _ := claimedPast7DaysL[cashMethodId]
		claimedPast1Day, _ := claimedPast1DayL[cashMethodId]
		cCtx = rfcontext.AppendParams(ctx, "moulding ", map[string]interface{}{
			"claimedPast7Days": claimedPast1Day,
			"claimedPast1Day":  claimedPast7Days,
		})

		maxClaimable, err := cash_method_promotion.FinalPossiblePayout(c, claimedPast7Days.Amount, claimedPast1Day.Amount, *cm.CashMethodPromotion, 0, true)
		if err != nil {
			cCtx = rfcontext.AppendError(cCtx, err, "FinalPossiblePayout")
			log.Println(rfcontext.Fmt(cCtx))
		}

		var cashMethodPromotion *serializer.CashMethodPromotion
		if !cm.HasCashMethodPromotion() {
			payoutRate := cm.CashMethodPromotion.PayoutRate

			floorApplicable := float64(cm.CashMethodPromotion.MinPayout) / 100

			cashMethodPromotion = &serializer.CashMethodPromotion{
				PayoutRate:                    payoutRate,
				MaxPromotionAmount:            float64(maxClaimable) / 100,
				MinAmountForPayout:            floorApplicable,
				DefaultOptionPromotionAmounts: nil,
			}
		}

		cashMethodR := serializer.BuildCashMethodWithPromotion(cm, cashMethodPromotion)
		cashMethodsR = append(cashMethodsR, cashMethodR)
	}

	var r serializer.Response
	if s.TopupOnly {
		r.Data = cashMethodsR
	} else {
		r.Data = util.MapSlice(cashMethodsR, func(cm serializer.CashMethod) serializer.CashMethod {
			firstTopup, err := model.FirstTopup(c, user.ID)
			if err != nil || len(firstTopup.ID) == 0 {
				cm.MinAmount = max(conf.GetCfg().WithdrawMinNoDeposit/100, cm.MinAmount)
			}
			return cm
		})
	}

	responseB, _ := json.Marshal(r)
	ctx = rfcontext.AppendDescription(ctx, fmt.Sprintf("response %s", string(responseB)))

	go log.Println(rfcontext.Fmt(ctx))
	return r, nil
}
