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
	vipRecordVipRuleId := vip.VipRule.ID

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
			"cash_method_id":    cm.ID,
			"cash_method":       cm,
			"cm.DefaultOptions": cm.DefaultOptions,
		})
		cashMethodId := cm.ID

		claimedPast7Days, _ := claimedPast7DaysL[cashMethodId]
		claimedPast1Day, _ := claimedPast1DayL[cashMethodId]
		cCtx = rfcontext.AppendParams(ctx, "moulding ", map[string]interface{}{
			"claimedPast7Days": claimedPast1Day,
			"claimedPast1Day":  claimedPast7Days,
		})

		var cashMethodPromotion *serializer.CashMethodPromotion

		if cm.HasCashMethodPromotion() {
			// for each option, individual query to get respective cashMethodPromotionOfSelection
			cashMethodDefaultOptions := cm.DefaultOptions
			selections := make([]serializer.DefaultCashMethodPromotionOption, 0, len(cashMethodDefaultOptions))

			var claimable int64 // for cashmethodpromotion.MaxPromotionAmount
			for idx, _selectionAmount := range cashMethodDefaultOptions {
				sCtx := rfcontext.Nonce(cCtx)
				selectionAmount := int64(_selectionAmount) // overcast

				cashMethodPromotionOfSelection, sErr := cash_method_promotion.ByCashMethodIdAndVipId(nil, cm.ID, vipRecordVipRuleId, nil, &selectionAmount)

				if sErr != nil {
					sCtx = rfcontext.AppendParams(sCtx, "cashMethodDefaultOption", map[string]interface{}{
						"selection_amount": selectionAmount,
					})
					sCtx = rfcontext.AppendError(sCtx, sErr, "ByCashMethodIdAndVipId")
				}

				{
					if idx == 0 { // assumes for cash method promotions wrt vip level and cash method, there's only 1 daily and 1 weekly max_payout => use any idx
						// top level claimable should account for user claimed amount
						_claimable, clErr := cash_method_promotion.FinalPossiblePayout(c, claimedPast7Days.Amount, claimedPast1Day.Amount, cashMethodPromotionOfSelection, nil)
						if clErr != nil {
							sCtx = rfcontext.AppendError(sCtx, clErr, "FinalPossiblePayout")
							log.Println(rfcontext.Fmt(sCtx))
						} else {
							claimable = _claimable
						}
					}
				}

				// QQ: extra百分比和“+XX“不會變 因为这个是display给全部人知道这个支付渠道有这个活动的 user达到了上限是那个user的问题 ，所以不会变
				_claimable, clErr := cash_method_promotion.FinalPossiblePayout(c, 0, 0, cashMethodPromotionOfSelection, nil)
				if clErr != nil {
					sCtx = rfcontext.AppendError(sCtx, clErr, "FinalPossiblePayout")
					log.Println(rfcontext.Fmt(sCtx))
				}

				label := fmt.Sprintf("%f", float64(selectionAmount)/100)

				//_maxClaimable = max(_maxClaimable, float64(_claimable))
				//maxPayoutRate = max(maxPayoutRate, cashMethodPromotionOfSelection.PayoutRate)

				selections = append(selections, serializer.DefaultCashMethodPromotionOption{
					SelectionAmount:     float64(selectionAmount) / 100,
					Label:               label,
					Icon:                serializer.Url("xxxx"),
					BonusRate:           cashMethodPromotionOfSelection.PayoutRate,
					BonusAmount:         float64(_claimable) / 100,
					NeedCustomerSupport: false,
				})
			}

			stats, rrerr := cash_method_promotion.ConfigStats(nil, &cashMethodId, &vipRecordVipRuleId, nil)
			if rrerr != nil {
				rfcontext.AppendErrorAsWarn(cCtx, rrerr, "ConfigStats")
				log.Println(rfcontext.Fmt(cCtx))
			}

			cashMethodPromotion = &serializer.CashMethodPromotion{
				PayoutRate:                        stats.PayoutRate_Max,
				MaxPromotionAmount:                float64(claimable) / 100,
				MinAmountForPayout:                float64(stats.MinPayout_Min) / 100,
				DefaultCashMethodPromotionOptions: selections,
			}
		}

		cashMethodR := serializer.BuildCashMethodWithCashMethodPromotion(cm, cashMethodPromotion)
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
