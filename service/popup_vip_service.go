package service

import (
	"web-api/model"
	"web-api/serializer"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type VipService struct {
}

func (service *VipService) Get(c *gin.Context) (r serializer.Response, err error) {
	u, _ := c.Get("user")
	user := u.(model.User)
	vip, err := model.GetVipWithDefault(nil, user.ID)
	if err != nil {
		return
	}
	currentVipRule := vip.VipRule
	data := map[string]int64{
		"vip_level": currentVipRule.VIPLevel,
	}

	r = serializer.Response{
		Data: data,
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
	return
}
