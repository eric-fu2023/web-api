package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"blgit.rfdev.tech/zhibo/utilities"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
)

type EmailOtpService struct {
	Email     string `form:"email" json:"email" binding:"required,email"`
	CheckUser bool   `form:"check_user" json:"check_user"`
	Action    string `form:"action" json:"action" binding:"required"`
}

func (service *EmailOtpService) GetEmail(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)

	service.Email = strings.ToLower(service.Email)

	if service.CheckUser {
		exists := service.checkExisting()
		if !exists {
			return serializer.ParamErr(c, service, i18n.T("account_invalid"), nil)
		}
	}

	otpSent, err := cache.GetOtp(c, service.Action, service.Email)
	if err != nil && errors.Is(err, cache.ErrInvalidOtpAction) {
		return serializer.ParamErr(c, service, i18n.T("invalid_otp_action"), nil)
	}
	if err != nil {
		return serializer.GeneralErr(c, err)
	}

	if otpSent != "" {
		return serializer.ParamErr(c, service, i18n.T("Email_wait"), nil)
	}

	var otp string
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 6; i++ {
		otp += strconv.Itoa(rand.Intn(9))
	}

	if os.Getenv("ENV") == "production" || os.Getenv("SEND_EMAIL_IN_TEST") == "true" {
		if err := service.sendEmail(c, otp); err != nil {
			return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("Email_fail"), err)
		}
	}

	err = cache.SetOtp(c, service.Action, service.Email, otp)
	if err != nil {
		return serializer.GeneralErr(c, err)
	}

	resp := serializer.SendOtp{}
	if os.Getenv("ENV") == "local" || os.Getenv("ENV") == "staging" {
		resp.Otp = otp
	}

	return serializer.Response{
		Msg:  i18n.T("success"),
		Data: resp,
	}
}

func (service *EmailOtpService) GetUsernameEmail(c *gin.Context, username string) serializer.Response {
	service.Email = strings.ToLower(service.Email)

	i18n := c.MustGet("i18n").(i18n.I18n)
	otpSent := cache.RedisSessionClient.Get(context.TODO(), "otp:"+username)
	if otpSent.Val() != "" {
		return serializer.ParamErr(c, service, i18n.T("Email_wait"), nil)
	}

	var otp string
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 6; i++ {
		otp += strconv.Itoa(rand.Intn(9))
	}

	if os.Getenv("ENV") == "production" || os.Getenv("SEND_EMAIL_IN_TEST") == "true" {
		if err := service.sendEmail(c, otp); err != nil {
			return serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("Email_fail"), err)
		}
	}
	cache.RedisSessionClient.Set(context.TODO(), "otp:"+username, otp, 2*time.Minute)

	resp := serializer.SendOtp{}
	if os.Getenv("ENV") == "local" || os.Getenv("ENV") == "staging" {
		resp.Otp = otp
	}

	return serializer.Response{
		Msg:  i18n.T("success"),
		Data: resp,
	}
}

func (service *EmailOtpService) sendEmail(c *gin.Context, otp string) error {
	i18n := c.MustGet("i18n").(i18n.I18n)

	emailProvider := utilities.EmailProvider{
		MailGunDomain:     os.Getenv("MAILGUN_DOMAIN"),
		MailGunPrivateKey: os.Getenv("MAILGUN_PRIVATE_KEY"),
		MailGunSender:     os.Getenv("MAILGUN_SENDER"),
	}
	if err := emailProvider.Send(service.Email, i18n.T("Otp_email_subject"), fmt.Sprintf(i18n.T("Otp_html_email"), otp)); err != nil {
		return err
	}

	event := model.OtpEvent{
		OtpEvent: ploutos.OtpEvent{
			Email:    service.Email,
			Otp:      otp,
			Provider: utilities.MailGunName,
			DateTime: time.Now().Format(time.DateTime),
			BrandId:  int64(c.GetInt("_brand")),
			Method:   ploutos.OtpEventMethodEmail,
		},
	}
	if err := model.LogOtpEvent(event); err != nil {
		// Just log error
		util.GetLoggerEntry(c).Errorf("Log otp event error: %s", err.Error())
	}

	return nil
}

func (service *EmailOtpService) checkExisting() bool {
	var user model.User
	emailHash := util.MobileEmailHash(service.Email)
	row := model.DB.Where(`email_hash`, emailHash).First(&user).RowsAffected
	if row > 0 {
		return true
	}
	return false
}
