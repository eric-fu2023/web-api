package common

import (
	constant "blgit.rfdev.tech/taya/captcha-go/const"
	"encoding/base64"
	"errors"
	"web-api/service/aj_captcha"
)

var ErrCaptchaInitialization = errors.New("captcha initialization error")

type Captcha struct {
	PointJson string `form:"point_json" json:"point_json"`
	Token     string `form:"token" json:"token"`
}

func CheckCaptcha(pointJson string, token string) (err error) {
	ajCaptcha := aj_captcha.AjCaptcha
	if ajCaptcha == nil {
		err = ErrCaptchaInitialization
		return
	}
	_, err = base64.StdEncoding.DecodeString(pointJson)
	if err != nil {
		return
	}
	err = ajCaptcha.GetService(constant.BlockPuzzleCaptcha).Verification(token, pointJson)
	if err != nil {
		return
	}
	return
}
