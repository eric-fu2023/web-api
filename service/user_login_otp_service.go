package service

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type UserOtpVerificationService struct {
	Otp         string `form:"otp" json:"otp" binding:"required"`
	CountryCode string `form:"country_code" json:"country_code"`
	Mobile      string `form:"mobile" json:"mobile"`
	Action      string `form:"action" json:"action" binding:"required"`
	Email       string `form:"email" json:"email"`
}

func (s UserOtpVerificationService) Verify(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)

	s.CountryCode = util.FormatCountryCode(s.CountryCode)
	s.Mobile = strings.TrimPrefix(s.Mobile, "0")
	s.Email = strings.ToLower(s.Email)

	var user model.User
	var err error
	u, exists := c.Get("user")
	if !exists {
		user, err = model.GetUserByMobileOrEmail(s.CountryCode, s.Mobile, s.Email)
		if err != nil && errors.Is(err, model.ErrCannotFindUser) {
			return serializer.ParamErr(c, s, i18n.T("account_invalid"), err)
		}
		if err != nil {
			return serializer.GeneralErr(c, err)
		}
	} else {
		user = u.(model.User)
	}

	userKeys := []string{string(user.Email), user.CountryCode + string(user.Mobile)}
	otp, err := cache.GetOtpByUserKeys(c, s.Action, userKeys)
	if err != nil && errors.Is(err, cache.ErrInvalidOtpAction) {
		return serializer.ParamErr(c, s, i18n.T("invalid_otp_action"), nil)
	}
	if err != nil {
		return serializer.GeneralErr(c, err)
	}

	if otp != s.Otp {
		return serializer.Err(c, s, serializer.CodeOtpInvalid, i18n.T("otp_invalid"), nil)
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
	Channel     string `form:"channel" json:"channel"`
}

func (service *UserLoginOtpService) Login(c *gin.Context) serializer.Response {
	service.Email = strings.ToLower(service.Email)
	service.Username = strings.TrimSpace(strings.ToLower(service.Username))

	service.CountryCode = util.FormatCountryCode(service.CountryCode)
	service.Mobile = strings.TrimPrefix(service.Mobile, "0")
	mobileHash := util.MobileEmailHash(service.Mobile)
	emailHash := util.MobileEmailHash(service.Email)

	i18n := c.MustGet("i18n").(i18n.I18n)

	var user model.User
	otpUserKey := ""
	if service.Email != "" {
		otpUserKey = service.Email
	} else if service.CountryCode != "" && service.Mobile != "" {
		otpUserKey = service.CountryCode + service.Mobile
	} else if service.Username != "" {
		otpUserKey = service.Username
	} else {
		return serializer.ParamErr(c, service, i18n.T("Both_cannot_be_empty"), nil)
	}
	if !((os.Getenv("ENV") == "local" || os.Getenv("ENV") == "staging") && service.Otp == "159357") { // for testing convenience
		otp, err := cache.GetOtp(c, consts.SmsOtpActionLogin, otpUserKey)
		if err != nil && errors.Is(err, cache.ErrInvalidOtpAction) {
			return serializer.ParamErr(c, service, i18n.T("invalid_otp_action"), nil)
		}
		if err != nil {
			return serializer.GeneralErr(c, err)
		}
		if otp != service.Otp {
			go LogFailedLogin(c, user, consts.AuthEventLoginMethod["otp"], service.Email, service.CountryCode, service.Mobile)
			return serializer.Err(c, service, serializer.CodeOtpInvalid, i18n.T("otp_invalid"), nil)
		}
	}

	deviceInfo, err := util.GetDeviceInfo(c)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GetDeviceInfo error: %s", err.Error())
		return serializer.ParamErr(c, service, i18n.T("invalid_device_info"), err)
	}

	q := model.DB
	if service.Email != "" {
		q = q.Where(`email_hash`, emailHash)
	} else if service.CountryCode != "" && service.Mobile != "" {
		q = q.Where(`country_code = ? AND mobile_hash = ?`, service.CountryCode, mobileHash)
	} else if service.Username != "" {
		q = q.Where(`username = ?`, service.Username)
	}
	if rows := q.Scopes(model.ByActiveNonStreamerUser).Find(&user).RowsAffected; rows == 0 {
		// new user
		user = model.User{
			User: ploutos.User{
				CountryCode:             service.CountryCode,
				Status:                  1,
				Role:                    1, // default role user
				RegistrationIp:          c.ClientIP(),
				RegistrationDeviceUuid:  deviceInfo.Uuid,
				ReferralWagerMultiplier: 1,
				Channel:                 service.Channel,
			},
			Email:  ploutos.EncryptedStr(service.Email),
			Mobile: ploutos.EncryptedStr(service.Mobile),
		}
		if service.Email != "" {
			user.EmailHash = emailHash
		}
		if service.Mobile != "" {
			user.MobileHash = mobileHash
		}
		//user.BrandId = int64(c.MustGet("_brand").(int))
		//user.AgentId = int64(c.MustGet("_agent").(int))
		genNickname(&user)
		err = user.CreateWithDB(model.DB)
		if err != nil {
			return serializer.DBErr(c, service, i18n.T("User_add_fail"), err)
		}
	}

	setupRequired := false
	if user.Username == "" {
		setupRequired = true
	}

	tokenString, err := ProcessUserLogin(c, user, consts.AuthEventLoginMethod["otp"], service.Email, service.CountryCode, service.Mobile)
	if err != nil && errors.Is(err, ErrTokenGeneration) {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("Error_token_generation"), err)
	} else if err != nil && errors.Is(err, util.ErrInvalidDeviceInfo) {
		util.GetLoggerEntry(c).Errorf("processUserLogin error: %s", err.Error())
		return serializer.ParamErr(c, service, i18n.T("invalid_device_info"), err)
	} else if err != nil {
		util.GetLoggerEntry(c).Errorf("processUserLogin error: %s", err.Error())
		return serializer.GeneralErr(c, err)
	}

	return serializer.Response{
		Data: map[string]interface{}{
			"token":          tokenString,
			"setup_required": setupRequired,
		},
	}
}

func ProcessUserLogin(c *gin.Context, user model.User, loginMethod int, inputtedEmail, inputtedCountryCode, inputtedMobile string) (string, error) {
	tokenString, err := user.GenToken()
	if err != nil {
		return "", ErrTokenGeneration
	}
	if timeout, e := strconv.Atoi(os.Getenv("SESSION_TIMEOUT")); e == nil {
		val := map[string]interface{}{
			"token":    tokenString,
			"password": serializer.UserSignature(user.ID),
		}
		cache.RedisSessionClient.HSet(context.TODO(), user.GetRedisSessionKey(), val)
		cache.RedisSessionClient.Expire(context.TODO(), user.GetRedisSessionKey(), time.Duration(timeout)*time.Minute)
	}

	loginTime := time.Now()
	err = user.UpdateLoginInfo(c, loginTime)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("UpdateLoginInfo error: %s", err.Error())
		return "", err
	}

	go LogSuccessfulLogin(c, user, loginTime, loginMethod, inputtedEmail, inputtedCountryCode, inputtedMobile)
	return tokenString, nil
}

func LogSuccessfulLogin(c *gin.Context, user model.User, loginTime time.Time, loginMethod int, inputtedEmail, inputtedCountryCode, inputtedMobile string) {
	deviceInfo, err := util.GetDeviceInfo(c)
	if err != nil {
		// Just log error if failed
		util.GetLoggerEntry(c).Errorf("Get device info error: %s", err.Error())
	}

	event := model.AuthEvent{
		AuthEvent: ploutos.AuthEvent{
			UserId:      user.ID,
			Type:        consts.AuthEventType["login"],
			Status:      consts.AuthEventStatus["successful"],
			DateTime:    loginTime.Format(time.DateTime),
			LoginMethod: loginMethod,
			Username:    user.Username,
			Email:       inputtedEmail,
			CountryCode: inputtedCountryCode,
			Mobile:      inputtedMobile,
			Ip:          c.ClientIP(),
			Platform:    deviceInfo.Platform,
			BrandId:     user.BrandId,
			AgentId:     user.AgentId,
			Uuid:        deviceInfo.Uuid,
		},
	}

	if err = model.LogAuthEvent(event); err != nil {
		util.GetLoggerEntry(c).Errorf("Log auth event error: %s", err.Error())
	}
}

func LogFailedLogin(c *gin.Context, user model.User, loginMethod int, inputtedEmail, inputtedCountryCode, inputtedMobile string) (err error) {
	deviceInfo, err := util.GetDeviceInfo(c)
	if err != nil {
		// Just log error if failed
		util.GetLoggerEntry(c).Errorf("Get device info error: %s", err.Error())
		return
	}

	event := model.AuthEvent{
		AuthEvent: ploutos.AuthEvent{
			Type:        consts.AuthEventType["login"],
			Status:      consts.AuthEventStatus["failed"],
			DateTime:    time.Now().Format(time.DateTime),
			LoginMethod: loginMethod,
			Username:    user.Username,
			Email:       inputtedEmail,
			CountryCode: inputtedCountryCode,
			Mobile:      inputtedMobile,
			Ip:          c.ClientIP(),
			Platform:    deviceInfo.Platform,
			//BrandId:     int64(c.MustGet("_brand").(int)),
			//AgentId:     int64(c.MustGet("_agent").(int)),
			Uuid: deviceInfo.Uuid,
		},
	}

	if err = model.LogAuthEvent(event); err != nil {
		util.GetLoggerEntry(c).Errorf("Log auth event error: %s", err.Error())
		return
	}
	return
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
