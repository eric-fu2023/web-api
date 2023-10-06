package service

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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
	if err := q.Scopes(model.ByActiveNonStreamerUser).First(&user).Error; err != nil {
		return serializer.DBErr(c, service, errStr, nil)
	}

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
		util.Log().Error("update last login ip and time err", err)
	}

	if err = service.logSuccessfulLogin(c, user, loginTime); err != nil {
		util.Log().Error("log successful login err", err)
	}

	return serializer.Response{
		Data: map[string]interface{}{
			"token": tokenString,
		},
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

	err := service.logFailedLogin(c, user)
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

func (service *UserLoginPasswordService) logSuccessfulLogin(c *gin.Context, user model.User, loginTime time.Time) error {
	deviceInfo, err := util.GetDeviceInfo(c)
	if err != nil {
		// Just log error if failed
		util.Log().Error("get device info err", err)
	}

	event := model.AuthEvent{
		AuthEventC: models.AuthEventC{
			UserId:      user.ID,
			Type:        consts.AuthEventType["login"],
			Status:      consts.AuthEventStatus["successful"],
			DateTime:    time.Now().Format(time.DateTime),
			LoginMethod: consts.AuthEventLoginMethod["password"],
			Email:       service.Email,
			CountryCode: service.CountryCode,
			Mobile:      service.Mobile,
			Username:    user.Username,
			Ip:          c.ClientIP(),
			Platform:    deviceInfo.Platform,
		},
	}

	return model.LogAuthEvent(event)
}

func (service *UserLoginPasswordService) logFailedLogin(c *gin.Context, user model.User) error {
	deviceInfo, err := util.GetDeviceInfo(c)
	if err != nil {
		// Just log error if failed
		util.Log().Error("get device info err", err)
	}

	event := model.AuthEvent{
		AuthEventC: models.AuthEventC{
			UserId:      user.ID,
			Type:        consts.AuthEventType["login"],
			Status:      consts.AuthEventStatus["failed"],
			DateTime:    time.Now().Format(time.DateTime),
			LoginMethod: consts.AuthEventLoginMethod["password"],
			Email:       service.Email,
			CountryCode: service.CountryCode,
			Mobile:      service.Mobile,
			Username:    user.Username,
			Ip:          c.ClientIP(),
			Platform:    deviceInfo.Platform,
		},
	}

	return model.LogAuthEvent(event)
}
