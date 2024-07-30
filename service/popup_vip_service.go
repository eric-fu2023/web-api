package service

import (
	"context"
	"fmt"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type VipService struct {
}

func (service *VipService) Get(c *gin.Context) (data map[string]int64, err error) {
	u, _ := c.Get("user")
	user := u.(model.User)
	vip, err := model.GetVipWithDefault(nil, user.ID)
	if err != nil {
		return
	}
	currentVipRule := vip.VipRule
	data = map[string]int64{
		"vip_level": currentVipRule.VIPLevel,
	}
	_, shown_err := service.Shown(c)
	if shown_err != nil {
		return data, shown_err
	}
	return
}

func (service *VipService) Shown(c *gin.Context) (r serializer.Response, err error) {
	u, _ := c.Get("user")
	user := u.(model.User)
	vip, err := model.GetVipWithDefault(nil, user.ID)
	if err != nil {
		return
	}
	currentVipRule := vip.VipRule

	PopupRecord := ploutos.PopupRecord{
		UserID:   user.ID,
		VipLevel: currentVipRule.VIPLevel,
		Type:     2,
	}
	err = model.DB.Create(&PopupRecord).Error
	
	var popup_service PopupService
	key := popup_service.buildKey(user.ID)
	res := cache.RedisClient.Set(context.Background(), key, "3", time.Hour*24)
	if res.Err() != nil{
		fmt.Print("insert vip popup record into redis failed ", key)
	}
	return
}
