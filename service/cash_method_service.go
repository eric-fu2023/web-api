package service

import (
	"web-api/conf"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"

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
	var list []model.CashMethod
	list, err = model.CashMethod{}.List(c, s.WithdrawOnly, s.TopupOnly, deviceInfo.Platform, brand)
	if err != nil {
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
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
		r.Data = util.MapSlice(list, serializer.BuildCashMethod)
	} else {
		r.Data = util.MapSlice(list, serializer.Modifier(serializer.BuildCashMethod, func(cm serializer.CashMethod) serializer.CashMethod {
			firstTopup, err := model.FirstTopup(c, user.ID)
			if err != nil || len(firstTopup.ID) == 0 {
				cm.MinAmount = conf.GetCfg().WithdrawMinNoDeposit / 100
			}
			return cm
		}))
	}
	return
}
