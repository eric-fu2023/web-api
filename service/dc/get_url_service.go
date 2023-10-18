package dc

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"crypto/md5"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"strings"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"
)

type GetUrlService struct {
	common.Platform
	GameId     int64 `form:"game_id" json:"game_id" binding:"required"`
	Fullscreen bool  `form:"fullscreen" json:"fullscreen"`
}

func (service *GetUrlService) Get(c *gin.Context) (res serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	if user.Username == "" {
		res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("finish_setup"), nil)
		return
	}

	tokenString := c.MustGet("_token_string").(string)
	sess := cache.RedisSessionClient.Get(context.TODO(), user.GetRedisSessionKey())
	if sess.Val() != tokenString {
		res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), nil)
		return
	}
	tokenHash := md5.Sum([]byte(tokenString))

	gpu, _, _, _, err := common.GetUserAndSum(consts.GameVendor["dc"], user.Username)
	if err != nil {
		var currency ploutos.CurrencyGameVendor
		err = model.DB.Where(`game_vendor_id`, consts.GameVendor["dc"]).Where(`currency_id`, user.CurrencyId).First(&currency).Error
		if err != nil {
			res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("empty_currency_id"), err)
			return
		}
		var game UserRegister
		err = game.CreateUser(user, currency.Value)
		if err != nil {
			res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("dc_create_user_failed"), err)
			return
		}
	}

	client := util.DCFactory.NewClient()
	lang := c.MustGet("_language").(string)
	if lang == "zh" {
		lang = "zh_hans"
	}
	cc := strings.ToUpper(c.MustGet("_country_code").(string))
	r, err := client.LoginGame(user.Username, hex.EncodeToString(tokenHash[:]), service.GameId, gpu.ExternalCurrency, lang, consts.PlatformIdToDcPlatformId[service.Platform.Platform], cc, "", &service.Fullscreen)
	if err != nil {
		res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	res = serializer.Response{
		Data: r,
	}
	return
}

type FunPlayService struct {
	GetUrlService
	CurrencyId int64 `form:"currency_id" json:"currency_id" binding:"required"`
}

func (service *FunPlayService) FunPlay(c *gin.Context) (res serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)

	var currency string
	err = model.DB.Model(ploutos.CurrencyGameVendor{}).Select(`value`).
		Where(`game_vendor_id`, consts.GameVendor["dc"]).Where(`currency_id`, service.CurrencyId).First(&currency).Error
	if err != nil {
		res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	client := util.DCFactory.NewClient()
	lang := c.MustGet("_language").(string)
	if lang == "zh" {
		lang = "zh_hans"
	}
	cc := strings.ToUpper(c.MustGet("_country_code").(string))
	r, err := client.TryGame(service.GameId, currency, lang, consts.PlatformIdToDcPlatformId[service.Platform.Platform], cc, "", &service.Fullscreen)
	if err != nil {
		res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	res = serializer.Response{
		Data: r,
	}
	return
}
