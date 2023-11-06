package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/service/fb"
	"web-api/service/saba"
	"web-api/util/i18n"
)

type UserSetPasswordService struct {
	CountryCode string `form:"country_code" json:"country_code" validate:"omitempty,startswith=+"`
	Mobile      string `form:"mobile" json:"mobile" validate:"omitempty,number"`
	Password    string `form:"password" json:"password" binding:"required,password"`
	Otp         string `form:"otp" json:"otp" binding:"required"`
}

func (service *UserSetPasswordService) SetPassword(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var user model.User
	u, isUser := c.Get("user")
	if isUser {
		user = u.(model.User)
	} else {
		if err := model.DB.Where(`country_code`, service.CountryCode).Where(`mobile`, service.Mobile).First(&user).Error; err != nil {
			return serializer.ParamErr(c, service, i18n.T("Mobile_invalid"), err)
		}
	}

	if user.Password != "" {
		if e := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(service.Password)); e == nil {
			return serializer.ParamErr(c, service, i18n.T("same_password"), nil)
		}
	}

	otp := cache.RedisSessionClient.Get(context.TODO(), "otp:"+user.Email)
	if otp.Err() == redis.Nil {
		otp = cache.RedisSessionClient.Get(context.TODO(), "otp:"+user.CountryCode+user.Mobile)
	}
	if otp.Val() != service.Otp {
		return serializer.ParamErr(c, service, i18n.T("otp_invalid"), nil)
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(service.Password), model.PassWordCost)
	if err != nil {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("password_encrypt_failed"), err)
	}

	if err = model.DB.Model(&user).Update(`password`, string(bytes)).Error; err != nil {
		return serializer.DBErr(c, service, i18n.T("password_update_failed"), err)
	}

	common.SendNotification(user.ID, consts.Notification_Type_Password_Reset, i18n.T("notification_password_reset_title"), i18n.T("notification_password_reset"))

	return serializer.Response{
		Msg: i18n.T("success"),
	}
}

type UserFinishSetupService struct {
	Username   string `form:"username" json:"username" binding:"required,username"`
	Password   string `form:"password" json:"password" binding:"required,password"`
	Code       string `form:"code" json:"code"`
	CurrencyId int64  `form:"currency_id" json:"currency_id" binding:"required,numeric"`
}

func (service *UserFinishSetupService) Set(c *gin.Context) serializer.Response {
	service.Username = strings.ToLower(service.Username)
	service.Code = strings.ToUpper(service.Code)

	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	if user.Username != "" && user.Password != "" && user.CurrencyId != 0 {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("setup_finished"), nil)
	}

	var existing model.User
	if r := model.DB.Unscoped().Where(`username`, service.Username).Limit(1).Find(&existing).RowsAffected; r != 0 {
		return serializer.Err(c, service, serializer.CodeExistingUsername, i18n.T("existing_username"), nil)
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(service.Password), model.PassWordCost)
	if err != nil {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("password_encrypt_failed"), err)
	}
	user.Username = service.Username
	user.Password = string(bytes)
	user.CurrencyId = service.CurrencyId
	user.BrandId = consts.DefaultBrand
	user.AgentId = consts.DefaultAgent

	if service.Code != "" {
		var agent ploutos.Agent
		if err = model.DB.Where(`code`, service.Code).First(&agent).Error; err == nil {
			user.BrandId = agent.BrandId
			user.AgentId = agent.ID
		} else {
			return serializer.ParamErr(c, service, i18n.T("invalid_agent"), nil)
		}
	}

	err = CreateUser(user)
	if err != nil && errors.Is(err, ErrEmptyCurrencyId) {
		return serializer.ParamErr(c, service, i18n.T("empty_currency_id"), nil)
	} else if err != nil && errors.Is(err, fb.ErrOthers) {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("fb_create_user_failed"), err)
	} else if err != nil && errors.Is(err, saba.ErrOthers) {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("saba_create_user_failed"), err)
	} else if err != nil {
		return serializer.DBErr(c, service, i18n.T("User_add_fail"), err)
	}

	common.SendNotification(user.ID, consts.Notification_Type_User_Registration, i18n.T("notification_welcome_title"), i18n.T("notification_welcome"))

	return serializer.Response{
		Msg: i18n.T("success"),
	}
}

type UserCheckUsernameService struct {
	Username string `form:"username" json:"username" binding:"required,excludesall=' '"`
}

func (service *UserCheckUsernameService) Check(c *gin.Context) serializer.Response {
	service.Username = strings.ToLower(service.Username)
	i18n := c.MustGet("i18n").(i18n.I18n)
	var existing model.User
	if r := model.DB.Where(`username`, service.Username).Limit(1).Find(&existing).RowsAffected; r != 0 {
		return serializer.Err(c, service, serializer.CodeExistingUsername, i18n.T("existing_username"), nil)
	}
	return serializer.Response{}
}

type UserCheckPasswordService struct {
	Password string `form:"password" json:"password" binding:"required"`
}

func (service *UserCheckPasswordService) Check(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)
	if user.Password != "" {
		if e := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(service.Password)); e == nil {
			return serializer.ParamErr(c, service, i18n.T("same_password"), nil)
		}
	}
	return serializer.Response{}
}
