package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
)

type AppConfigService struct {
	Platform
	Key string `form:"key" json:"key"`
}

func (service *AppConfigService) Get(c *gin.Context) (r serializer.Response, err error) {
	var configs []ploutos.AppConfig
	brand := c.MustGet(`_brand`).(int)
	agent := c.MustGet(`_agent`).(int)
	if err = model.DB.Scopes(model.ByBrandAgentDeviceAndKey(int64(brand), int64(agent), service.Platform.Platform, service.Key)).Find(&configs).Error; err == nil {
		cf := make(map[string]map[string]string)
		for _, b := range configs {
			_, exists := cf[b.Name]
			if !exists {
				cf[b.Name] = make(map[string]string)
			}
			cf[b.Name][b.Key] = b.Value
		}
		r = serializer.Response{
			Data: cf,
		}
	}
	return
}
