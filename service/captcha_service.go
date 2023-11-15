package service

import (
	constant "blgit.rfdev.tech/taya/captcha-go/const"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"web-api/serializer"
	"web-api/service/aj_captcha"
)

type CaptchaGetService struct {
}

func (service *CaptchaGetService) Get(c *gin.Context) (r map[string]interface{}, err error) {
	ajCaptcha := aj_captcha.AjCaptcha
	if ajCaptcha == nil {
		return
	}
	data, err := ajCaptcha.GetService(constant.BlockPuzzleCaptcha).Get()
	if err != nil {
		return
	}
	r = map[string]interface{}{
		"repCode": "0000",
		"repData": map[string]interface{}{
			"originalImageBase64": data["originalImageBase64"],
			"jigsawImageBase64":   data["jigsawImageBase64"],
			"token":               data["token"],
			"secretKey":           data["secretKey"],
			//"result": false,
			//"opAdmin": false,
		},
		"success": true,
		"error":   false,
	}
	return
}

type CaptchaCheckService struct {
	CaptchaType string `form:"captchaType" json:"captchaType"`
	ClientUid   string `form:"clientUid" json:"clientUid"`
	PointJson   string `form:"pointJson" json:"pointJson"`
	Token       string `form:"token" json:"token"`
	Ts          int64  `form:"ts" json:"ts"`
}

func (service *CaptchaCheckService) Check(c *gin.Context) (r serializer.Response, err error) {
	ajCaptcha := aj_captcha.AjCaptcha
	if ajCaptcha == nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, "Something went wrong.", nil)
		return
	}
	_, err = base64.StdEncoding.DecodeString(service.PointJson)
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeCaptchaInvalid, "Captcha is invalid", err)
		return
	}
	err = ajCaptcha.GetService(constant.BlockPuzzleCaptcha).Check(service.Token, service.PointJson)
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeCaptchaInvalid, "Captcha is invalid", err)
		return
	}

	r = serializer.Response{
		Data: map[string]string{
			"point_json": service.PointJson,
			"token":      service.Token,
		},
	}
	return
}
