package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"strings"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type UserLoginOtpService struct {
	CountryCode string `form:"country_code" json:"country_code"`
	Mobile      string `form:"mobile" json:"mobile"`
	Email       string `form:"email" json:"email"`
	Otp         string `form:"otp" json:"otp" binding:"required"`
	Username    string `form:"username" json:"username"`
	Password    string `form:"password" json:"password"`
}

func (service *UserLoginOtpService) Login(c *gin.Context) serializer.Response {
	service.Email = strings.ToLower(service.Email)

	i18n := c.MustGet("i18n").(i18n.I18n)

	var user model.User
	errStr := ""
	key := "otp:"
	if service.Email != "" {
		key += service.Email
		errStr = i18n.T("Email_invalid")
	} else if service.CountryCode != "" && service.Mobile != "" {
		key += service.CountryCode + service.Mobile
		errStr = i18n.T("Mobile_invalid")
	} else {
		return serializer.ParamErr(i18n.T("Both_cannot_be_empty"), nil)
	}
	otp := cache.RedisSessionClient.Get(context.TODO(), key)
	if otp.Val() != service.Otp {
		return serializer.ParamErr(errStr, nil)
	}

	if rows := model.DB.Where(`email`, service.Email).Or(`country_code = ? AND mobile = ?`, service.CountryCode, service.Mobile).Find(&user).RowsAffected; rows == 0 { // new user
		if service.Username == "" || service.Password == "" {
			return serializer.ParamErr(i18n.T("empty_username_password"), nil)
		}
		var existing model.User
		if r := model.DB.Where(`username`, service.Username).Limit(1).Find(&existing).RowsAffected; r != 0 {
			return serializer.Err(serializer.CodeExistingUsername, i18n.T("existing_username"), nil)
		}
		bytes, err := bcrypt.GenerateFromPassword([]byte(service.Password), model.PassWordCost)
		if err != nil {
			return serializer.ParamErr(i18n.T("密码加密失败"), err)
		}
		user = model.User{
			Email:       service.Email,
			CountryCode: service.CountryCode,
			Mobile:      service.Mobile,
			Username:    service.Username,
			Password:    string(bytes),
			Status:      1,
			Role:        1, // default role user
		}
		user.BrandId = int64(c.MustGet("_brand").(int))
		user.AgentId = int64(c.MustGet("_agent").(int))
		if err := model.DB.Create(&user).Error; err != nil {
			return serializer.ParamErr(i18n.T("User_add_fail"), err)
		}
	}

	tokenString, err := user.GenToken()
	if err != nil {
		return serializer.ParamErr(i18n.T("Error_token_generation"), err)
	}
	cache.RedisSessionClient.Set(context.TODO(), strconv.Itoa(int(user.ID)), tokenString, 20*time.Minute)

	return serializer.Response{
		Data: map[string]interface{}{
			"token": tokenString,
		},
	}
}
