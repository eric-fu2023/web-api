package backend

import (
	"encoding/json"
	"os"

	"web-api/conf/consts"
	"web-api/model"
	"web-api/model/avatar"
	"web-api/serializer"
	"web-api/service"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type GetTokenService struct {
	Username   string `form:"username" json:"username" binding:"required,username"`
	BrandId    int64  `form:"brand_id" json:"brand_id" binding:"required"`
	CurrencyId int64  `form:"currency_id" json:"currency_id" binding:"required"`
	Nickname   string `form:"nickname" json:"nickname" binding:"required"`
	Pin        string `form:"pin" json:"pin"`
	Code       string `form:"code" json:"code"`
	Ip         string `form:"ip" json:"ip"`
	DeviceUuid string `form:"device_uuid" json:"device_uuid"`
	Platform   string `form:"platform" json:"platform"`
	Channel    string `form:"channel" json:"channel"`
}

func (s *GetTokenService) Get(c *gin.Context) (r serializer.Response, err error) {
	var user model.User
	rows := model.DB.Scopes(ploutos.ByActiveNonStreamerUser).Where(`username`, s.Username).Where(`brand_id`, s.BrandId).Find(&user).RowsAffected
	if rows == 0 { // new user
		if s.CurrencyId == 0 {
			r = serializer.Err(c, s, serializer.CodeParamErr, "currency_id is required for user registration", nil)
			return
		}
		if s.Nickname == "" {
			r = serializer.Err(c, s, serializer.CodeParamErr, "nickname is required for user registration", nil)
			return
		}
		var agent ploutos.Agent
		if s.Code == "" {
			err = model.DB.Where(`brand_id`, s.BrandId).Order(`id`).First(&agent).Error
			if err != nil {
				r = serializer.Err(c, s, serializer.CodeDBError, "adding user failed: default agent not found", err)
				return
			}
		} else {
			err = model.DB.Where(`code`, s.Code).First(&agent).Error
			if err != nil {
				r = serializer.Err(c, s, serializer.CodeDBError, "adding user failed: agent not found", err)
				return
			}
		}
		user = model.User{
			User: ploutos.User{
				Username:                s.Username,
				BrandId:                 s.BrandId,
				AgentId:                 agent.ID,
				CurrencyId:              s.CurrencyId,
				Nickname:                s.Nickname,
				Status:                  1,
				Role:                    1,
				RegistrationIp:          s.Ip,
				RegistrationDeviceUuid:  s.DeviceUuid,
				ReferralWagerMultiplier: 1,
				Locale:                  os.Getenv("PLATFORM_LANGUAGE"),
			},
		}
		user.Avatar = avatar.GetRandomAvatarUrl()
		if s.Pin != "" {
			bytes, err := bcrypt.GenerateFromPassword([]byte(s.Pin), model.PassWordCost)
			if err != nil {
				util.GetLoggerEntry(c).Errorf("GenerateFromPassword error: %s", err.Error())
			}
			user.SecondaryPassword = string(bytes)
		}
		err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
			if len(s.Channel) > 0 {
				user.Channel = s.Channel
				service.ConnectChannelAgent(&user, tx)
			}
			err = user.CreateWithDB(model.DB)
			if err != nil {
				util.GetLoggerEntry(c).Errorf("CreateWithDB error: %s", err.Error())
				return
			}
			err = service.CreateUser(&user)
			if err != nil {
				util.GetLoggerEntry(c).Errorf("CreateUser error: %s", err.Error())
			}
			return
		})
		if err != nil {
			r = serializer.Err(c, s, serializer.CodeDBError, "adding new user failed", err)
			return
		}
	}
	deviceInfo := map[string]string{
		"uuid":     s.DeviceUuid,
		"platform": s.Platform,
	}
	j, err := json.Marshal(deviceInfo)
	if err != nil {
		r = serializer.Err(c, s, serializer.CodeGeneralError, "adding new user failed", err)
		return
	}
	c.Request.Header.Add("Device-Info", string(j)) // ProcessUserLogin and other functions get Device-Info from the header
	c.Request.Header.Add("X-Forwarded-For", s.Ip)
	token, err := service.ProcessUserLogin(c, user, consts.AuthEventLoginMethod["backend"], "", "", "")
	if err != nil {
		r = serializer.Err(c, s, serializer.CodeDBError, "adding new user failed", err)
		return
	}

	r = serializer.Response{
		Data: map[string]interface{}{
			"token": token,
		},
	}
	return
}

type PinService struct {
	Username string `form:"username" json:"username" binding:"required"`
	BrandId  int64  `form:"brand_id" json:"brand_id" binding:"required"`
	Pin      string `form:"pin" json:"pin" binding:"required"`
}

func (s *PinService) Set(c *gin.Context) (r serializer.Response, err error) {
	var user model.User
	err = model.DB.Scopes(ploutos.ByActiveNonStreamerUser).Where(`username`, s.Username).Where(`brand_id`, s.BrandId).First(&user).Error
	if err != nil {
		r = serializer.Err(c, s, serializer.CodeDBError, "user not found", err)
		return
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(s.Pin), model.PassWordCost)
	if err != nil {
		util.GetLoggerEntry(c).Errorf("GenerateFromPassword error: %s", err.Error())
	}
	err = model.DB.Model(&user).Update(`secondary_password`, string(bytes)).Error
	if err != nil {
		r = serializer.Err(c, s, serializer.CodeDBError, "pin update failed", err)
		return
	}
	r = serializer.Response{
		Msg: "success",
	}
	return
}
