package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strings"
	"time"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

const (
	passwordLockEndTimeKey = "password_lock_end_time:%d" // password_lock_end_time:<userId>
)

var (
	errNoEmailOrMobile = errors.New("user has no email or mobile")
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
	service.Username = strings.TrimSpace(strings.ToLower(service.Username))

	mobileHash := serializer.MobileEmailHash(service.Mobile)
	emailHash := serializer.MobileEmailHash(service.Email)

	i18n := c.MustGet("i18n").(i18n.I18n)

	var user model.User
	q := model.DB
	errStr := ""
	if service.Username != "" {
		q = q.Where(`username`, service.Username)
		errStr = i18n.T("Username_password_invalid")
	} else if service.Email != "" {
		q = q.Where(`email_hash`, emailHash)
		errStr = i18n.T("Email_password_invalid")
	} else if service.CountryCode != "" && service.Mobile != "" {
		q = q.Where(`country_code`, service.CountryCode).Where(`mobile_hash`, mobileHash)
		errStr = i18n.T("Mobile_password_invalid")
	} else {
		return serializer.ParamErr(c, service, i18n.T("Both_cannot_be_empty"), nil)
	}
	if err := q.Scopes(model.ByActiveNonStreamerUser).First(&user).Error; err != nil {
		return serializer.DBErr(c, service, errStr, nil)
	}

	var otpResp serializer.Response
	isAccountLocked, err := service.checkAccountLock(user)
	if err != nil {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("Error_password_lock"), nil)
	}
	if isAccountLocked {
		return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("Password_lock_wait"), nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(service.Password)); err != nil {
		return service.handlePasswordMismatch(c, user)
	}

	respData := map[string]interface{}{}
	if os.Getenv("PASSWORD_LOGIN_REQUIRES_OTP") == "true" {
		otpResp, err = service.sendOtp(c, user)
		if errors.Is(err, errNoEmailOrMobile) {
			return serializer.ParamErr(c, service, i18n.T("User_needs_email_or_mobile"), nil)
		} else if err != nil {
			return serializer.GeneralErr(c, err)
		} else if otpResp.Code != 0 {
			return otpResp
		}

		// Return masked email and mobile
		if service.Email != "" || user.Email != "" {
			respData["email"] = util.MaskEmail(string(user.Email))
		} else if (service.CountryCode != "" && service.Mobile != "") || (user.CountryCode != "" && user.Mobile != "") {
			respData["country_code"] = user.CountryCode
			respData["mobile"] = util.MaskMobile(string(user.Mobile))
		}

		if os.Getenv("ENV") == "local" || os.Getenv("ENV") == "staging" {
			if otpRespData, ok := otpResp.Data.(serializer.SendOtp); ok {
				respData["otp"] = otpRespData.Otp
			}
		}
	} else {
		tokenString, err := ProcessUserLogin(c, user, consts.AuthEventLoginMethod["password"], service.Email, service.CountryCode, service.Mobile)
		if err != nil && errors.Is(err, ErrTokenGeneration) {
			return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("Error_token_generation"), err)
		} else if err != nil && errors.Is(err, util.ErrInvalidDeviceInfo) {
			util.GetLoggerEntry(c).Errorf("processUserLogin error: %s", err.Error())
			return serializer.ParamErr(c, service, i18n.T("invalid_device_info"), err)
		} else if err != nil {
			util.GetLoggerEntry(c).Errorf("processUserLogin error: %s", err.Error())
			return serializer.GeneralErr(c, err)
		}

		respData["token"] = tokenString
	}

	return serializer.Response{
		Msg:  i18n.T("success"),
		Data: respData,
	}
}

func (service *UserLoginPasswordService) checkAccountLock(user model.User) (bool, error) {
	key := fmt.Sprintf(passwordLockEndTimeKey, user.ID)
	res := cache.RedisSessionClient.Get(context.TODO(), key)
	if res.Val() == "" {
		return false, nil
	}

	lockEndTime, err := time.ParseInLocation(time.DateTime, res.Val(), time.Local)
	if err != nil {
		util.Log().Error("account lock time parse err", err)
		return false, err
	}

	if time.Now().After(lockEndTime) {
		return false, nil
	}
	return true, nil
}

func (service *UserLoginPasswordService) handlePasswordMismatch(c *gin.Context, user model.User) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	res := serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("login_failed"), nil)

	err := LogFailedLogin(c, user, consts.AuthEventLoginMethod["password"], service.Email, service.CountryCode, service.Mobile)
	if err != nil {
		util.Log().Error("log failed login err", err)
		return res
	}

	latestEvents, err := model.GetLatestAuthEvents(user.ID, 5)
	if err != nil {
		util.Log().Error("get latest auth events err", err)
		return res
	}

	failedCount := 0
	for _, event := range latestEvents {
		if event.Type == consts.AuthEventType["login"] &&
			event.LoginMethod == consts.AuthEventLoginMethod["password"] &&
			event.Status == consts.AuthEventStatus["failed"] {
			failedCount++
		}
	}
	if failedCount < 5 {
		return res
	}

	latestTime, err := time.ParseInLocation(time.DateTime, latestEvents[0].DateTime, time.Local)
	if err != nil {
		util.Log().Error("parse auth event latest time err", err, latestEvents[0].DateTime)
		return res
	}
	earliestTime, err := time.ParseInLocation(time.DateTime, latestEvents[4].DateTime, time.Local)
	if err != nil {
		util.Log().Error("parse auth event earliest time err", err, latestEvents[4].DateTime)
		return res
	}
	if timeDifference := latestTime.Sub(earliestTime); timeDifference > 30*time.Minute {
		return res
	}

	key := fmt.Sprintf(passwordLockEndTimeKey, user.ID)
	thirtyMinsLater := time.Now().Add(30 * time.Minute).Format(time.DateTime)
	status := cache.RedisSessionClient.Set(context.TODO(), key, thirtyMinsLater, 30*time.Minute)
	if status.Err() != nil {
		util.Log().Error("set password lock time in redis err", err, thirtyMinsLater)
		return res
	}

	return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("Password_lock_wait"), nil)
}

func (service *UserLoginPasswordService) sendOtp(c *gin.Context, user model.User) (serializer.Response, error) {
	var resp serializer.Response
	emailOtpService := EmailOtpService{Email: string(user.Email), Action: consts.SmsOtpActionLogin}
	smsOtpService := SmsOtpService{CountryCode: user.CountryCode, Mobile: string(user.Mobile), Action: consts.SmsOtpActionLogin}

	if service.Email != "" {
		resp = emailOtpService.GetEmail(c)
	} else if service.CountryCode != "" && service.Mobile != "" {
		resp = smsOtpService.GetSMS(c)
	} else if user.Email != "" {
		resp = emailOtpService.GetUsernameEmail(c, user.Username)
	} else if user.CountryCode != "" && user.Mobile != "" {
		resp = smsOtpService.GetUsernameSMS(c, user.Username)
	} else {
		return serializer.Response{}, errNoEmailOrMobile
	}

	return resp, nil
}
