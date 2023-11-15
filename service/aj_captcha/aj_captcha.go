package aj_captcha

import (
	"blgit.rfdev.tech/taya/captcha-go/config"
	constant "blgit.rfdev.tech/taya/captcha-go/const"
	"blgit.rfdev.tech/taya/captcha-go/service"
	"image/color"
	"os"
	"strconv"
)

var AjCaptcha *service.CaptchaServiceFactory

func Init() {
	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		return
	}
	cfg := config.NewConfig()
	cfg.CacheType = constant.RedisCacheKey
	cfg.Watermark = &config.WatermarkConfig{Color: color.RGBA{R: 255, G: 255, B: 255, A: 255}, FontSize: 12, Text: os.Getenv("APP_NAME")}
	cfg.ResourcePath = "./captcha"
	AjCaptcha = service.NewCaptchaServiceFactory(cfg)
	AjCaptcha.RegisterCache(constant.RedisCacheKey, service.NewConfigRedisCacheService([]string{os.Getenv("REDIS_ADDR")}, "", "", false, db))
	//AjCaptcha.RegisterService(constant.ClickWordCaptcha, service.NewClickWordCaptchaService(AjCaptcha))
	AjCaptcha.RegisterService(constant.BlockPuzzleCaptcha, service.NewBlockPuzzleCaptchaService(AjCaptcha))
}
