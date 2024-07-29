package service

import (
	"web-api/conf"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type CasheMethodListService struct {
	WithdrawOnly bool `form:"withdraw_only" json:"withdraw_only"`
	TopupOnly    bool `form:"topup_only" json:"topup_only"`
}

func (s CasheMethodListService) List(c *gin.Context) (r serializer.Response, err error) {
	brand := c.MustGet(`_brand`).(int)

	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user, _ := u.(model.User)
	deviceInfo, _ := util.GetDeviceInfo(c)

	vip, err := model.GetVipWithDefault(c, user.ID)
	if err != nil {
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	var list []model.CashMethod
	list, err = model.CashMethod{}.List(c, s.WithdrawOnly, s.TopupOnly, deviceInfo.Platform, brand, int(vip.VipRule.ID))
	if err != nil {
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	weeklyAmountRecords, dailyAmountRecords, err := model.GetWeeklyAndDailyCashMethodPromotionRecord(c, 0, user.ID)
	maxPromotionAmountByCashMethodId := map[int64]int64{}
	util.MapSlice(list, func(a model.CashMethod) (err error) {
		if a.CashMethodPromotion == nil {
			return
		}
		weeklyAmount := util.FindOrDefault(weeklyAmountRecords, func(b models.CashMethodPromotionRecord) bool {
			return b.CashMethodId == a.ID
		}).Amount
		dailyAmount := util.FindOrDefault(dailyAmountRecords, func(b models.CashMethodPromotionRecord) bool {
			return b.CashMethodId == a.ID
		}).Amount

		maxAmount, err := model.GetMaxCashMethodPromotionAmount(c, weeklyAmount, dailyAmount, *a.CashMethodPromotion, user.ID, 0, true)
		if err != nil {
			util.GetLoggerEntry(c).Error("HandleCashMethodPromotion GetMaxAmountPayment", err)
		}
		maxPromotionAmountByCashMethodId[a.ID] = maxAmount
		return
	})

	if s.TopupOnly {
		// var firstTime bool
		// firstTime, err = model.CashOrder{}.IsFirstTime(c, user.ID)
		// if err != nil {
		// 	r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		// 	return
		// }
		// minAmount := conf.GetCfg().FirstTopupMinimum / 100
		// if !firstTime && loggedIn {
		// 	minAmount = conf.GetCfg().TopupMinimum / 100
		// }
		r.Data = util.MapSlice(list, func(a model.CashMethod) serializer.CashMethod {
			return serializer.BuildCashMethod(a, maxPromotionAmountByCashMethodId)
		})
	} else {
		r.Data = util.MapSlice(list, serializer.Modifier(
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
	return
}
