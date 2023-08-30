package service

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

type UserSetPasswordService struct {
	Password string `form:"password" json:"password" binding:"required"`
	Otp      string `form:"otp" json:"otp" binding:"required"`
}

func (service *UserSetPasswordService) SetPassword(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	otp := cache.RedisSessionClient.Get(context.TODO(), "otp:"+user.Email)
	if otp.Err() == redis.Nil {
		otp = cache.RedisSessionClient.Get(context.TODO(), "otp:"+user.CountryCode+user.Mobile)
	}
	if otp.Val() != service.Otp {
		return serializer.ParamErr(c, service, i18n.T("验证码错误"), nil)
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(service.Password), model.PassWordCost)
	if err != nil {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("密码加密失败"), err)
	}

	if err = model.DB.Model(&user).Update(`password`, string(bytes)).Error; err != nil {
		return serializer.DBErr(c, service, i18n.T("密码修改失败"), err)
	}

	return serializer.Response{
		Msg: i18n.T("success"),
	}
}

type UserFinishSetupService struct {
	Username    string `form:"username" json:"username" binding:"required"`
	Password    string `form:"password" json:"password" binding:"required"`
	CurrencyId  int64 `form:"currency_id" json:"currency_id" binding:"required,numeric"`
}

func (service *UserFinishSetupService) Set(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	if user.Username != "" || user.Password != "" || user.CurrencyId != 0 {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("setup_finished"), nil)
	}

	var existing model.User
	if r := model.DB.Where(`username`, service.Username).Limit(1).Find(&existing).RowsAffected; r != 0 {
		return serializer.Err(c, service, serializer.CodeExistingUsername, i18n.T("existing_username"), nil)
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(service.Password), model.PassWordCost)
	if err != nil {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("密码加密失败"), err)
	}
	user.Username = service.Username
	user.Password = string(bytes)
	user.CurrencyId = service.CurrencyId
	tx := model.DB.Begin()
	err = tx.Save(&user).Error
	if err != nil {
		tx.Rollback()
		return serializer.DBErr(c, service, i18n.T("User_add_fail"), err)
	}

	userSum := model.UserSum{
		models.UserSumC{
			UserId: user.ID,
		},
	}
	err = tx.Create(&userSum).Error
	if err != nil {
		tx.Rollback()
		return serializer.DBErr(c, service, i18n.T("User_add_fail"), err)
	}

	var currency model.CurrencyGameProvider
	err = model.DB.Where(`game_provider_id`, consts.GameProvider["fb"]).Where(`currency_id`, service.CurrencyId).First(&currency).Error
	if err != nil {
		tx.Rollback()
		return serializer.ParamErr(c, service, i18n.T("empty_currency_id"), nil)
	}
	client := util.FBFactory.NewClient()
	res, err := client.CreateUser(user.Username, []int64{}, 0)
	if err != nil {
		tx.Rollback()
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("fb_create_user_failed"), err)
	}
	gpu := model.GameProviderUser{
		models.GameProviderUserC{
			GameProviderId:     consts.GameProvider["fb"],
			UserId:             user.ID,
			ExternalUserId:     user.Username,
			ExternalCurrencyId: currency.Value,
			ExternalId:         fmt.Sprintf("%d", res),
		},
	}
	err = tx.Save(&gpu).Error
	if err != nil {
		tx.Rollback()
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("fb_create_user_failed"), err)
	}
	tx.Commit()

	return serializer.Response{
		Msg: i18n.T("success"),
	}
}
