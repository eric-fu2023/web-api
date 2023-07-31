package service

import (
	"github.com/gin-gonic/gin"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type UserDeleteService struct {
	Otp string `form:"otp" json:"otp" binding:"required"`
}

func (service *UserDeleteService) Delete(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	if err := model.DB.Where(`sms_otp = ? AND sms_otp_expired_at > ?`, service.Otp, time.Now().Format("2006-01-02 15:04:05")).Or(`email_otp = ? AND email_otp_expired_at > ?`, service.Otp, time.Now().Format("2006-01-02 15:04:05")).First(&user).Error; err != nil {
		return serializer.ParamErr(i18n.T("验证码错误"), err)
	}
	if rows := model.DB.Delete(&user).RowsAffected; rows < 1 {
		return serializer.ParamErr(i18n.T("失败"), nil)
	}
	return serializer.Response{
		Msg: i18n.T("success"),
	}
}
