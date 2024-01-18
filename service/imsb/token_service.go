package imsb

import (
	"blgit.rfdev.tech/taya/game-service/imsb/callback"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"
)

var (
	ImsbErrInvalidMemberCode = errors.New("invalid member code")
)

type TokenService struct {
}

func (service *TokenService) Get(c *gin.Context) (res serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	if user.Username == "" {
		res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("finish_setup"), nil)
		return
	}

	_, _, _, _, err = common.GetUserAndSum(model.DB, consts.GameVendor["imsb"], user.Username)
	var currency ploutos.CurrencyGameVendor
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = model.DB.Where(`game_vendor_id`, consts.GameVendor["imsb"]).Where(`currency_id`, user.CurrencyId).First(&currency).Error
		if err != nil {
			res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("empty_currency_id"), err)
			return
		}
		var game UserRegister
		err = game.CreateUser(user, currency.Value)
		if err != nil {
			res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("imsb_create_user_failed"), err)
			return
		}
	}

	token, err := util.AesEncrypt([]byte(fmt.Sprintf(`%d`, user.ID)))
	if err != nil {
		res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	b64Token := base64.StdEncoding.EncodeToString([]byte(token))

	res = serializer.Response{
		Data: b64Token,
	}
	return
}

type ValidateTokenService struct {
	Token string `form:"token" json:"token" binding:"required"`
}

func (service *ValidateTokenService) Validate(c *gin.Context) (res callback.ValidateToken, err error) {
	token, err := base64.StdEncoding.DecodeString(service.Token)
	if err != nil {
		return
	}
	uidStr, err := util.AesDecrypt(string(token))
	if err != nil {
		return
	}
	userId, err := strconv.Atoi(uidStr)
	if err != nil {
		return
	}
	var user model.User
	err = model.DB.Where(`id`, userId).Where(`status`, 1).Where(`username != ''`).First(&user).Error
	if err != nil {
		err = ImsbErrInvalidMemberCode
		return
	}
	gpu, _, _, _, err := common.GetUserAndSum(model.DB, consts.GameVendor["imsb"], user.Username)
	if err != nil {
		return
	}

	res = callback.ValidateToken{
		BaseResponse: callback.BaseResponse{
			StatusCode: 100,
			StatusDesc: "Success",
		},
		MemberCode: user.Username,
		Currency:   gpu.ExternalCurrency,
		IpAddress:  user.LastLoginIp,
	}
	return
}
