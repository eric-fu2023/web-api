package fb

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"fmt"
	"github.com/gin-gonic/gin"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service"
	"web-api/util"
	"web-api/util/i18n"
)

type TokenService struct {
	service.Platform
}

func (service *TokenService) Get(c *gin.Context) (res serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	if user.Username == "" {
		res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("finish_setup"), nil)
		return
	}

	client := util.FBFactory.NewClient()
	r, err := client.GetToken(user.Username, consts.PlatformIdToFbPlatformId[service.Platform.Platform], "")
	if err != nil {
		var currency ploutos.CurrencyGameProvider
		err = model.DB.Where(`game_provider_id`, consts.GameProvider["fb"]).Where(`currency_id`, user.CurrencyId).First(&currency).Error
		if err != nil {
			res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("empty_currency_id"), err)
			return
		}
		var extId int64
		extId, err = client.CreateUser(user.Username, []int64{}, 0)
		if err != nil {
			res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("fb_create_user_failed"), err)
			return
		}
		fbGpu := ploutos.GameProviderUser{
			ploutos.GameProviderUserC{
				GameProviderId:     consts.GameProvider["fb"],
				UserId:             user.ID,
				ExternalUserId:     user.Username,
				ExternalCurrencyId: currency.Value,
				ExternalId:         fmt.Sprintf("%d", extId),
			},
		}
		err = model.DB.Save(&fbGpu).Error
		if err != nil {
			res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("fb_create_user_failed"), err)
			return
		}
		r, err = client.GetToken(user.Username, consts.PlatformIdToFbPlatformId[service.Platform.Platform], "")
		if err != nil {
			res = serializer.Err(c, service, serializer.CodeGeneralError, "", err)
			return
		}
	}
	res = serializer.Response{
		Data: r,
	}
	return
}
