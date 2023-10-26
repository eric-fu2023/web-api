package service

import (
	"context"
	"math/rand"
	"strings"
	"time"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type UserOtpVerificationService struct {
	Otp string `form:"otp" json:"otp" binding:"required"`
}

func (s UserOtpVerificationService) Verify(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	user := c.MustGet("user").(model.User)

	key := "otp:" + user.Email
	otp := cache.RedisSessionClient.Get(context.TODO(), key)
	if otp.Err() == redis.Nil {
		key = "otp:" + user.CountryCode + user.Mobile
		otp = cache.RedisSessionClient.Get(context.TODO(), key)
	}
	if otp.Val() != s.Otp {
		return serializer.ParamErr(c, s, i18n.T("验证码错误"), nil)
	}
	// THINK: may not need this
	// _ = cache.RedisSessionClient.Expire(context.TODO(), key, 2*time.Minute)
	return serializer.Response{
		Msg: i18n.T("success"),
	}
}

type UserLoginOtpService struct {
	CountryCode string `form:"country_code" json:"country_code"`
	Mobile      string `form:"mobile" json:"mobile"`
	Email       string `form:"email" json:"email"`
	Username    string `form:"username" json:"username"`
	Otp         string `form:"otp" json:"otp" binding:"required"`
}

func (service *UserLoginOtpService) Login(c *gin.Context) serializer.Response {
	service.Email = strings.ToLower(service.Email)
	service.Username = strings.ToLower(service.Username)

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
	} else if service.Username != "" {
		key += service.Username
		errStr = i18n.T("Username_invalid")
	} else {
		return serializer.ParamErr(c, service, i18n.T("Both_cannot_be_empty"), nil)
	}
	otp := cache.RedisSessionClient.Get(context.TODO(), key)
	if otp.Val() != service.Otp {
		return serializer.ParamErr(c, service, errStr, nil)
	}

	q := model.DB
	if service.Email != "" {
		q = q.Where(`email`, service.Email)
	} else if service.CountryCode != "" && service.Mobile != "" {
		q = q.Where(`country_code = ? AND mobile = ?`, service.CountryCode, service.Mobile)
	} else if service.Username != "" {
		q = q.Where(`username = ?`, service.Username)
	}
	if rows := q.Scopes(model.ByActiveNonStreamerUser).Find(&user).RowsAffected; rows == 0 { // new user
		user = model.User{
			UserC: models.UserC{
				Email:       service.Email,
				CountryCode: service.CountryCode,
				Mobile:      service.Mobile,
				Status:      1,
				Role:        1, // default role user
			},
		}
		user.BrandId = int64(c.MustGet("_brand").(int))
		user.AgentId = int64(c.MustGet("_agent").(int))
		genNickname(&user)
		err := model.DB.Create(&user).Error
		if err != nil {
			return serializer.DBErr(c, service, i18n.T("User_add_fail"), err)
		}
	}

	setupRequired := false
	if user.Username == "" {
		setupRequired = true
	}

	tokenString, err := user.GenToken()
	if err != nil {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("Error_token_generation"), err)
	}
	cache.RedisSessionClient.Set(context.TODO(), user.GetRedisSessionKey(), tokenString, 20*time.Minute)

	loginTime := time.Now()
	update := model.User{
		UserC: models.UserC{
			LastLoginIp:   c.ClientIP(),
			LastLoginTime: loginTime,
		},
	}
	if err = model.DB.Model(&user).
		Select("last_login_ip", "last_login_time").
		Updates(update).Error; err != nil {
		util.GetLoggerEntry(c).Errorf("Update last login ip and time error: %s", err.Error())
	}

	if err = service.logSuccessfulLogin(c, user, loginTime); err != nil {
		util.GetLoggerEntry(c).Errorf("Log successful login error: %s", err.Error())
	}

	return serializer.Response{
		Data: map[string]interface{}{
			"token":          tokenString,
			"setup_required": setupRequired,
		},
	}
}

func (service *UserLoginOtpService) logSuccessfulLogin(c *gin.Context, user model.User, loginTime time.Time) error {
	deviceInfo, err := util.GetDeviceInfo(c)
	if err != nil {
		// Just log error if failed
		util.GetLoggerEntry(c).Errorf("Get device info error: %s", err.Error())
	}

	event := model.AuthEvent{
		AuthEventC: models.AuthEventC{
			UserId:      user.ID,
			Type:        consts.AuthEventType["login"],
			Status:      consts.AuthEventStatus["successful"],
			DateTime:    loginTime.Format(time.DateTime),
			LoginMethod: consts.AuthEventLoginMethod["otp"],
			Username:    user.Username,
			Email:       service.Email,
			CountryCode: service.CountryCode,
			Mobile:      service.Mobile,
			Ip:          c.ClientIP(),
			Platform:    deviceInfo.Platform,
			BrandId:     user.BrandId,
			AgentId:     user.AgentId,
		},
	}

	return model.LogAuthEvent(event)
}

func genNickname(user *model.User) {
	var nicks []map[string]interface{}
	model.DB.Table(`nicknames`).Find(&nicks)
	if len(nicks) > 0 {
		rand.Seed(time.Now().UnixNano())
		r1 := rand.Intn(len(nicks))
		r2 := rand.Intn(len(nicks))
		user.Nickname = nicks[r1]["first_name"].(string) + nicks[r2]["last_name"].(string)
	}
}
