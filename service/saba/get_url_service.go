package saba

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"os"
	"strconv"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service"
	"web-api/util"
	"web-api/util/i18n"
)

type GetUrlService struct {
	service.Platform
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
			curr, e := strconv.Atoi(currency.Value)
			if e != nil {
				res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("saba_create_user_failed"), err)
				return
			}
			if r, e := client.CreateMember(user.Username, int64(curr), os.Getenv("GAME_SABA_ODDS_TYPE")); e == nil {
				sabaGpu := ploutos.GameVendorUser{
					ploutos.GameVendorUserC{
						GameVendorId:     consts.GameVendor["saba"],
						UserId:           user.ID,
						ExternalUserId:   user.Username,
						ExternalCurrency: currency.Value,
						ExternalId:       r,
					},
				}
				err = model.DB.Save(&sabaGpu).Error
				if err != nil {
					res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("saba_create_user_failed"), err)
					return
				}
			} else {
				res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("saba_create_user_failed"), e)
				return
			}
			url, err = client.GetSabaUrl(user.Username, consts.PlatformIdToSabaPlatformId[service.Platform.Platform])
			if err != nil {
				res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("saba_user_error"), err)
				return
			}
		} else {
			res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("saba_user_error"), err)
			return
		}
	}
	res = serializer.Response{
		Data: url,
	}
	return
}
