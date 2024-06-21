package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service"
	"web-api/util"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
)

func Me(c *gin.Context) {
	var service service.MeService
	if err := c.ShouldBind(&service); err == nil {
		res := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserLogout(c *gin.Context) {
	i18n := c.MustGet("i18n").(i18n.I18n)

	u, _ := c.Get("user")
	user := u.(model.User)
	cmd := cache.RedisSessionClient.Del(context.TODO(), user.GetRedisSessionKey())
	if cmd.Err() == redis.Nil {
		c.JSON(401, serializer.Response{
			Code: serializer.CodeCheckLogin,
			Msg:  i18n.T("account_invalid"),
		})
		c.Abort()
		return
	}

	deviceInfo, err := util.GetDeviceInfo(c)
	if err != nil {
		// Just log error if failed
		util.Log().Error("get device info err", err)
	}

	// Add logout log
	event := model.AuthEvent{
		AuthEvent: ploutos.AuthEvent{
			UserId:   user.ID,
			Username: user.Username,
			Type:     consts.AuthEventType["logout"],
			Status:   consts.AuthEventStatus["successful"],
			DateTime: time.Now().Format(time.DateTime),
			Ip:       c.ClientIP(),
			Platform: deviceInfo.Platform,
			BrandId:  user.BrandId,
			AgentId:  user.AgentId,
			Uuid:     deviceInfo.Uuid,
		},
	}
	if err := model.LogAuthEvent(event); err != nil {
		// Just log error if failed
		util.Log().Error("log logout auth event err", err)
	}

	c.JSON(200, serializer.Response{
		Code: 0,
		Msg:  i18n.T("logout_success"),
	})
}

func SmsOtp(c *gin.Context) {
	var service service.SmsOtpService
	if err := c.ShouldBindWith(&service, binding.Form); err == nil {
		res := service.GetSMS(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func EmailOtp(c *gin.Context) {
	var service service.EmailOtpService
	if err := c.ShouldBindWith(&service, binding.Form); err == nil {
		res := service.GetEmail(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func WhatsAppOtp(c *gin.Context) {
	var service service.WhatsAppOtpService
	if err := c.ShouldBindWith(&service, binding.Form); err == nil {
		res := service.GetWhatsApp(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserLoginOtp(c *gin.Context) {
	var service service.UserLoginOtpService
	if err := c.ShouldBindWith(&service, binding.Form); err == nil {
		res := service.Login(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserLoginPassword(c *gin.Context) {
	var service service.UserLoginPasswordService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res := service.Login(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserFinishSetup(c *gin.Context) {
	var service service.UserFinishSetupService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res := service.Set(c)
		c.JSON(200, res)
	} else {
		if usernameValidationErrorWithMsg(c, service, err, "validation_username") {
			return
		}
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserCheckUsername(c *gin.Context) {
	var service service.UserCheckUsernameService
	if err := c.ShouldBind(&service); err == nil {
		res := service.Check(c)
		c.JSON(200, res)
	} else {
		if usernameValidationErrorWithMsg(c, service, err, "validation_username") {
			return
		}
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserCheckPassword(c *gin.Context) {
	var service service.UserCheckPasswordService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res := service.Check(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserSetPassword(c *gin.Context) {
	var service service.UserSetPasswordService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res := service.SetPassword(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserDelete(c *gin.Context) {
	var service service.UserDeleteService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res := service.Delete(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func ProfileUpdate(c *gin.Context) {
	var service service.ProfileUpdateService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res := service.Update(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func NicknameUpdate(c *gin.Context) {
	var service service.NicknameUpdateService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res := service.Update(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func ProfilePicUpload(c *gin.Context) {
	var service service.ProfilePicService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res := service.Upload(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserSetSecondaryPassword(c *gin.Context) {
	var service service.UserSecondaryPasswordService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res := service.SetSecondaryPassword(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserSetMobile(c *gin.Context) {
	var service service.UserSetMobileService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res := service.Set(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserSetEmail(c *gin.Context) {
	var service service.UserSetEmailService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res := service.Set(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserCounters(c *gin.Context) {
	var service service.CounterService
	if err := c.ShouldBind(&service); err == nil {
		res := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserWallets(c *gin.Context) {
	var service service.WalletService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserSyncWallet(c *gin.Context) {
	var service service.SyncWalletService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Update(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserRecallFund(c *gin.Context) {
	var service service.RecallFundService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Recall(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func InternalRecallFund(c *gin.Context) {
	var service service.InternalRecallFundService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Recall(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

var ErrEmptyMobileNumber = errors.New("mobile number required but empty")
var ErrExtraFieldMobileNumber = errors.New("mobile should be unset")

func UserRegister(requireMobile bool, bypassSetMobileOtpVerify bool) func(*gin.Context) {
	return func(c *gin.Context) {
		var service service.UserRegisterService
		if err := c.ShouldBindWith(&service, binding.Form); err != nil {
			if usernameValidationErrorWithMsg(c, service, err, "invalid_username") {
				return
			}
			c.JSON(400, ErrorResponse(c, service, err))
			return
		}

		i18n := c.MustGet("i18n").(i18n.I18n)
		if requireMobile != bypassSetMobileOtpVerify {
			// storing mobile in db implies mobile has gone through verification.
			// hence if mobile is required, otp verification bypass must explicitly be enabled
			// the converse is true. if no bypass => user should not be required to, and provide, mobile
			util.Log().Error("$v", errors.New("invalid server config"))
			c.JSON(http.StatusNotImplemented, ErrorResponseWithMsg(c, service, errors.New("not Implemented"), "Not Implemented"))
			return
		}
		if requireMobile {
			if service.Mobile == "" || service.CountryCode == "" {
				c.JSON(400, ErrorResponseWithMsg(c, service, ErrEmptyMobileNumber, i18n.T("Mobile_invalid")))
				return
			}
		} else if !(service.Mobile == "" && service.CountryCode == "") {
			c.JSON(400, ErrorResponseWithMsg(c, service, ErrExtraFieldMobileNumber, i18n.T("Mobile_invalid")))
			return
		}

		res := service.Register(c, bypassSetMobileOtpVerify)
		c.JSON(200, res)
	}
}

func usernameValidationErrorWithMsg(c *gin.Context, service any, err error, i18nKey string) (exists bool) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		i18n := c.MustGet("i18n").(i18n.I18n)
		for _, fe := range ve {
			if fe.Field() == "Username" {
				c.JSON(400, ErrorResponseWithMsg(c, service, err, i18n.T(i18nKey)))
				exists = true
				return
			}
		}
	}
	return
}
