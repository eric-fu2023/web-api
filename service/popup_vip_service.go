package service

import (
	"context"
	"fmt"
	"os"
	"strconv"
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
		"vip_id":vip.VipID,
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
	
	key := "popup/records/" + time.Now().Format("2006-01-02")
	res := cache.RedisClient.HSet(context.Background(), key, user.ID, "2")
	expire_time, err := strconv.Atoi(os.Getenv("POPUP_RECORD_EXPIRE_MINS"))
	cache.RedisClient.ExpireNX(context.Background(), key, time.Duration(expire_time) * time.Minute)
	if res.Err() != nil{
		fmt.Print("insert vip popup record into redis failed ", key)
	}
	return

}
