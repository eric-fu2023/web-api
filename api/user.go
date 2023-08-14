package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"strconv"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/service"
	"web-api/util/i18n"
)

func UserLogout(c *gin.Context) {
	i18n := c.MustGet("i18n").(i18n.I18n)

	u, _ := c.Get("user")
	user := u.(model.User)
	cmd := cache.RedisSessionClient.Del(context.TODO(), strconv.Itoa(int(user.ID)))
	if cmd.Err() == redis.Nil {
		c.JSON(401, serializer.Response{
			Code:  serializer.CodeCheckLogin,
			Msg:   i18n.T("账号错误"),
		})
		c.Abort()
		return
	}

	c.JSON(200, serializer.Response{
		Code: 0,
		Msg:  i18n.T("登出成功"),
	})
}

func SmsOtp(c *gin.Context) {
	var service service.SmsOtpService
	if err := c.ShouldBind(&service); err == nil {
		res := service.GetSMS(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func EmailOtp(c *gin.Context) {
	var service service.EmailOtpService
	if err := c.ShouldBind(&service); err == nil {
		res := service.GetEmail(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserLoginOtp(c *gin.Context) {
	var service service.UserLoginOtpService
	if err := c.ShouldBind(&service); err == nil {
		res := service.Login(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserLoginPassword(c *gin.Context) {
	var service service.UserLoginPasswordService
	if err := c.ShouldBind(&service); err == nil {
		res := service.Login(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserSetPassword(c *gin.Context) {
	var service service.UserSetPasswordService
	if err := c.ShouldBind(&service); err == nil {
		res := service.SetPassword(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserDelete(c *gin.Context) {
	var service service.UserDeleteService
	if err := c.ShouldBind(&service); err == nil {
		res := service.Delete(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
