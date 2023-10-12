package saba

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"
)

type GetUrlService struct {
	common.Platform
}

func (service *GetUrlService) Get(c *gin.Context) (res serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var user model.User
	u, isUser := c.Get("user")
	if isUser {
		user = u.(model.User)
	}
	client := util.SabaFactory.NewClient()
	url, err := client.GetSabaUrl(user.Username, consts.PlatformIdToSabaPlatformId[service.Platform.Platform])
	if err != nil {
		if err.Error() == "member not found" {
			var currency ploutos.CurrencyGameVendor
			err = model.DB.Where(`game_vendor_id`, consts.GameVendor["saba"]).Where(`currency_id`, user.CurrencyId).First(&currency).Error
			if err != nil {
				res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("empty_currency_id"), err)
				return
			}
			var game UserRegister
			err = game.CreateUser(user, currency.Value)
			if err != nil {
				res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("saba_create_user_failed"), err)
				return
			}
			url, err = client.GetSabaUrl(user.Username, consts.PlatformIdToSabaPlatformId[service.Platform.Platform])
			if err != nil {
				res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("saba_user_error"), err)
				return
			}
		} else {
			res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
			return
		}
	}
	res = serializer.Response{
		Data: url,
	}
	return
}
