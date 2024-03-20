package backend

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
	"web-api/service"
)

type GetTokenService struct {
	Username   string `form:"username" json:"username" binding:"required"`
	BrandId    int64  `form:"brand_id" json:"brand_id" binding:"required"`
	CurrencyId int64  `form:"currency_id" json:"currency_id"`
	Nickname   string `form:"nickname" json:"nickname"`
	Code       string `form:"code" json:"code"`
	Ip         string `form:"ip" json:"ip"`
	DeviceUuid string `form:"device_uuid" json:"device_uuid"`
	Platform   string `form:"platform" json:"platform"`
}

func (s *GetTokenService) Get(c *gin.Context) (r serializer.Response, err error) {
	var user model.User
	rows := model.DB.Scopes(model.ByActiveNonStreamerUser).Where(`username`, s.Username).Where(`brand_id`, s.BrandId).Find(&user).RowsAffected
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
				Username:               s.Username,
				BrandId:                s.BrandId,
				AgentId:                agent.ID,
				CurrencyId:             s.CurrencyId,
				Nickname:               s.Nickname,
				Status:                 1,
				Role:                   1,
				RegistrationIp:         s.Ip,
				RegistrationDeviceUuid: s.DeviceUuid,
			},
		}
		err = service.CreateUser(&user)
		if err != nil {
			r = serializer.Err(c, s, serializer.CodeDBError, "adding new user failed", err)
			return
		}
	}
	loginService := service.UserLoginOtpService{}
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
	token, err := loginService.ProcessUserLogin(c, user)
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
