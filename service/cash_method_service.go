package service

import (
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"

	"github.com/gin-gonic/gin"
)

type CasheMethodListService struct {
	Platform     int64 `form:"platform" json:"platform"`
	WithdrawOnly bool  `form:"withdraw_only" json:"withdraw_only"`
	TopupOnly    bool  `form:"topup_only" json:"topup_only"`
}

func (s CasheMethodListService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var list []model.CashMethod
	list,err = model.CashMethod{}.List(c,s.WithdrawOnly,s.TopupOnly,s.Platform)
	if err != nil {
		r = serializer.Err(c, s, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	r.Data = util.MapSlice(list, serializer.BuildCashMethod)
	return
}