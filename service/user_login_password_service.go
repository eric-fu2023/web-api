package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type UserLoginPasswordService struct {
	CountryCode string `form:"country_code" json:"country_code"`
	Mobile      string `form:"mobile" json:"mobile"`
	Username    string `form:"username" json:"username"`
	Email       string `form:"email" json:"email"`
	Password    string `form:"password" json:"password" binding:"required"`
}

func (service *UserLoginPasswordService) Login(c *gin.Context) serializer.Response {
	service.Email = strings.ToLower(service.Email)

	i18n := c.MustGet("i18n").(i18n.I18n)

	var user model.User
	q := model.DB
	errStr := ""
	if service.Username != "" {
		q = q.Where(`username`, service.Username)
		errStr = i18n.T("Username_invalid")
	} else if service.Email != "" {
		q = q.Where(`email`, service.Email)
		errStr = i18n.T("Email_invalid")
	} else if service.CountryCode != "" && service.Mobile != "" {
		q = q.Where(`country_code`, service.CountryCode).Where(`mobile`, service.Mobile)
		errStr = i18n.T("Mobile_invalid")
	} else {
		return serializer.ParamErr(c, service, i18n.T("Both_cannot_be_empty"), nil)
	}
	if err := q.First(&user).Error; err != nil {
		return serializer.DBErr(c, service, errStr, nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(service.Password)); err != nil {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("login_failed"), nil)
	}

	tokenString, err := user.GenToken()
	if err != nil {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("Error_token_generation"), err)
	}
	cache.RedisSessionClient.Set(context.TODO(), user.GetRedisSessionKey(), tokenString, 20*time.Minute)

	return serializer.Response{
		Data: map[string]interface{}{
			"token": tokenString,
		},
	}
}
